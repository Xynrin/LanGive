package transfer

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// HeaderOffset 断点续传偏移头：客户端将从该字节位置开始追加
	HeaderOffset = "X-LanGive-Offset"
	// HeaderFileName 文件名头，避免 multipart
	HeaderFileName = "X-LanGive-Filename"
	// HeaderTotalSize 总大小头
	HeaderTotalSize = "X-LanGive-Total-Size"
	// HeaderTransferID 发送方生成的 transferID，服务端用它命名 .part 文件
	HeaderTransferID = "X-LanGive-Transfer-Id"
	// HeaderToken 一次性鉴权令牌
	HeaderToken = "X-LanGive-Token"

	partSuffix = ".part"
	chunkSize  = 64 * 1024
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type TransferStatus struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"` // "send" or "receive"
	FileName  string  `json:"file_name"`
	TotalSize int64   `json:"total_size"`
	SentSize  int64   `json:"sent_size"`
	Progress  float64 `json:"progress"`
	Status    string  `json:"status"` // "pending", "transferring", "completed", "failed", "cancelled"
	Error     string  `json:"error,omitempty"`
	PeerAddr  string  `json:"peer_addr"`
}

type Service struct {
	downloadPath string
	port         int
	server       *http.Server
	router       *gin.Engine

	transfers    map[string]*TransferStatus
	transfersMux sync.RWMutex

	// cancels 跟踪正在进行的发送任务，用于取消
	cancels    map[string]context.CancelFunc
	cancelsMux sync.Mutex

	clients    map[string]*websocket.Conn
	clientsMux sync.RWMutex

	// pendingRequests 接收端待用户确认的传入请求
	pendingRequests    map[string]*IncomingRequest
	pendingRequestsMux sync.Mutex

	// validTokens 用户已批准的一次性 token
	validTokens    map[string]*tokenInfo
	validTokensMux sync.Mutex

	// onIncomingRequest 收到新传输请求时回调（前端弹窗）
	onIncomingRequest func(*IncomingRequest)

	ctx    context.Context
	cancel context.CancelFunc
}

// IncomingRequest 接收端等待用户确认的传输请求
type IncomingRequest struct {
	ID         string `json:"id"`
	FromName   string `json:"from_name"`
	FromAddr   string `json:"from_addr"`
	FileName   string `json:"file_name"`
	TotalSize  int64  `json:"total_size"`
	ReceivedAt int64  `json:"received_at"`
}

type tokenInfo struct {
	transferID string
	expiresAt  time.Time
}

func NewService(downloadPath string, port int) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	return &Service{
		downloadPath:    downloadPath,
		port:            port,
		router:          router,
		transfers:       make(map[string]*TransferStatus),
		cancels:         make(map[string]context.CancelFunc),
		clients:         make(map[string]*websocket.Conn),
		pendingRequests: make(map[string]*IncomingRequest),
		validTokens:     make(map[string]*tokenInfo),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// SetOnIncomingRequest 注册前端弹窗回调
func (s *Service) SetOnIncomingRequest(fn func(*IncomingRequest)) {
	s.onIncomingRequest = fn
}

func (s *Service) Start() error {
	s.setupRoutes()

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()
	return nil
}

func (s *Service) Stop() {
	s.cancel()
	if s.server != nil {
		s.server.Shutdown(context.Background())
	}
}

func (s *Service) setupRoutes() {
	s.router.GET("/ws", s.handleWebSocket)
	s.router.GET("/resume", s.handleResume)
	s.router.POST("/upload", s.handleUpload)
	s.router.GET("/transfers", s.handleGetTransfers)
	s.router.POST("/cancel/:id", s.handleCancelTransfer)
	s.router.POST("/transfer/request", s.handleTransferRequest)
}

// handleTransferRequest 接收端处理对方发起的"请求传输"
// 创建一个待确认 IncomingRequest，回调前端弹窗，由前端调用 ApproveIncoming/RejectIncoming
func (s *Service) handleTransferRequest(c *gin.Context) {
	var req struct {
		FromName  string `json:"from_name"`
		FileName  string `json:"file_name"`
		TotalSize int64  `json:"total_size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	id := uuid.New().String()
	ir := &IncomingRequest{
		ID:         id,
		FromName:   req.FromName,
		FromAddr:   c.ClientIP(),
		FileName:   filepath.Base(req.FileName),
		TotalSize:  req.TotalSize,
		ReceivedAt: time.Now().Unix(),
	}
	s.pendingRequestsMux.Lock()
	s.pendingRequests[id] = ir
	s.pendingRequestsMux.Unlock()
	if s.onIncomingRequest != nil {
		go s.onIncomingRequest(ir)
	}

	// 阻塞等待用户在前端做出选择，最长 60s
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		s.validTokensMux.Lock()
		var token string
		var found bool
		for tk, ti := range s.validTokens {
			if ti.transferID == id {
				token = tk
				found = true
				break
			}
		}
		s.validTokensMux.Unlock()
		if found {
			c.JSON(http.StatusOK, gin.H{"approved": true, "token": token, "id": id})
			return
		}
		s.pendingRequestsMux.Lock()
		_, stillPending := s.pendingRequests[id]
		s.pendingRequestsMux.Unlock()
		if !stillPending {
			c.JSON(http.StatusForbidden, gin.H{"approved": false})
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
	// 超时按拒绝处理
	s.pendingRequestsMux.Lock()
	delete(s.pendingRequests, id)
	s.pendingRequestsMux.Unlock()
	c.JSON(http.StatusRequestTimeout, gin.H{"approved": false})
}

// ApproveIncoming 由前端调用，批准一个待传请求并发放 token
func (s *Service) ApproveIncoming(id string) (string, error) {
	s.pendingRequestsMux.Lock()
	ir, ok := s.pendingRequests[id]
	if ok {
		delete(s.pendingRequests, id)
	}
	s.pendingRequestsMux.Unlock()
	if !ok {
		return "", fmt.Errorf("request not found or already handled")
	}
	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	s.validTokensMux.Lock()
	s.validTokens[token] = &tokenInfo{transferID: ir.ID, expiresAt: time.Now().Add(5 * time.Minute)}
	s.validTokensMux.Unlock()
	return token, nil
}

// RejectIncoming 由前端调用，拒绝一个待传请求
func (s *Service) RejectIncoming(id string) error {
	s.pendingRequestsMux.Lock()
	defer s.pendingRequestsMux.Unlock()
	if _, ok := s.pendingRequests[id]; !ok {
		return fmt.Errorf("request not found")
	}
	delete(s.pendingRequests, id)
	return nil
}

// PendingRequests 返回当前待确认的传输请求快照
func (s *Service) PendingRequests() []*IncomingRequest {
	s.pendingRequestsMux.Lock()
	defer s.pendingRequestsMux.Unlock()
	out := make([]*IncomingRequest, 0, len(s.pendingRequests))
	for _, r := range s.pendingRequests {
		out = append(out, r)
	}
	return out
}

// consumeToken 校验并消耗 token，返回是否合法
func (s *Service) consumeToken(token string) bool {
	if token == "" {
		return false
	}
	s.validTokensMux.Lock()
	defer s.validTokensMux.Unlock()
	ti, ok := s.validTokens[token]
	if !ok {
		return false
	}
	if time.Now().After(ti.expiresAt) {
		delete(s.validTokens, token)
		return false
	}
	delete(s.validTokens, token)
	return true
}

func (s *Service) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	clientID := uuid.New().String()
	s.clientsMux.Lock()
	s.clients[clientID] = conn
	s.clientsMux.Unlock()
	defer func() {
		s.clientsMux.Lock()
		delete(s.clients, clientID)
		s.clientsMux.Unlock()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

// handleResume 返回服务端 .part 已写入的字节数，供发送端断点续传
// 必须带 name + size + tid（发送方 transferID）三个参数：
//   - finalPath 已存在且大小与 size 完全一致 → completed:true
//   - finalPath 已存在但大小不一致 → completed:false, offset:0（避免同名错跳）
//   - <name>.<tid>.part 存在 → 返回该 part 大小作为 offset
func (s *Service) handleResume(c *gin.Context) {
	name := filepath.Base(c.Query("name"))
	if name == "" || name == "." || name == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing name"})
		return
	}
	size, _ := strconv.ParseInt(c.Query("size"), 10, 64)
	tid := c.Query("tid")
	finalPath := filepath.Join(s.downloadPath, name)

	if st, err := os.Stat(finalPath); err == nil {
		if size > 0 && st.Size() == size {
			c.JSON(http.StatusOK, gin.H{"offset": st.Size(), "completed": true})
			return
		}
		// 同名但大小不一致：不复用，强制从头传（接收端会自动改名避免覆盖）
		c.JSON(http.StatusOK, gin.H{"offset": 0, "completed": false})
		return
	}
	if tid != "" {
		partPath := filepath.Join(s.downloadPath, name+"."+tid+partSuffix)
		if st, err := os.Stat(partPath); err == nil {
			c.JSON(http.StatusOK, gin.H{"offset": st.Size(), "completed": false})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"offset": 0, "completed": false})
}

// handleUpload 接收上传字节流
// 头部约定：
//   X-LanGive-Filename: 目标文件名
//   X-LanGive-Total-Size: 文件总大小
//   X-LanGive-Offset: 本次请求体的起始偏移（启用断点续传时设置）
func (s *Service) handleUpload(c *gin.Context) {
	if err := os.MkdirAll(s.downloadPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// A9：上传必须带已批准的 token
	if !s.consumeToken(c.GetHeader(HeaderToken)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
		return
	}

	name := filepath.Base(c.GetHeader(HeaderFileName))
	if name == "" || name == "." {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing filename header"})
		return
	}
	total, _ := strconv.ParseInt(c.GetHeader(HeaderTotalSize), 10, 64)
	offset, _ := strconv.ParseInt(c.GetHeader(HeaderOffset), 10, 64)
	tid := c.GetHeader(HeaderTransferID)
	if tid == "" {
		tid = uuid.New().String()
	}

	partPath := filepath.Join(s.downloadPath, name+"."+tid+partSuffix)
	finalPath := filepath.Join(s.downloadPath, name)

	// 校验本地已收字节是否与发送端声明的偏移一致
	var existing int64
	if st, err := os.Stat(partPath); err == nil {
		existing = st.Size()
	}
	if offset != existing {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "offset mismatch",
			"offset": existing,
		})
		return
	}

	dst, err := os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer dst.Close()

	transferID := tid
	status := &TransferStatus{
		ID:        transferID,
		Type:      "receive",
		FileName:  name,
		TotalSize: total,
		SentSize:  existing,
		Status:    "transferring",
		PeerAddr:  c.ClientIP(),
	}
	if total > 0 {
		status.Progress = float64(existing) / float64(total) * 100
	}
	s.transfersMux.Lock()
	s.transfers[transferID] = status
	s.transfersMux.Unlock()
	s.broadcastProgress(status)

	written := existing
	buf := make([]byte, chunkSize)
	for {
		nr, rerr := c.Request.Body.Read(buf)
		if nr > 0 {
			nw, werr := dst.Write(buf[:nr])
			written += int64(nw)
			status.SentSize = written
			if total > 0 {
				status.Progress = float64(written) / float64(total) * 100
			}
			s.broadcastProgress(status)
			if werr != nil {
				status.Status = "failed"
				status.Error = werr.Error()
				c.JSON(http.StatusInternalServerError, gin.H{"error": werr.Error()})
				return
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			// 客户端断开 → part 文件保留，后续可续传
			status.Status = "failed"
			status.Error = rerr.Error()
			s.broadcastProgress(status)
			c.JSON(http.StatusInternalServerError, gin.H{"error": rerr.Error()})
			return
		}
	}

	// 全部接收完毕：去掉 .part 后缀，自动避开同名文件
	if total == 0 || written >= total {
		dst.Close()
		target := uniquePath(finalPath)
		if err := os.Rename(partPath, target); err != nil {
			status.Status = "failed"
			status.Error = err.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		status.FileName = filepath.Base(target)
		status.Status = "completed"
		status.Progress = 100
	} else {
		status.Status = "paused"
	}
	s.broadcastProgress(status)
	c.JSON(http.StatusOK, gin.H{"id": transferID, "received": written, "status": status.Status, "name": status.FileName})
}

func (s *Service) handleGetTransfers(c *gin.Context) {
	c.JSON(http.StatusOK, s.GetTransfers())
}

func (s *Service) handleCancelTransfer(c *gin.Context) {
	id := c.Param("id")
	if err := s.CancelTransfer(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

func (s *Service) broadcastProgress(status *TransferStatus) {
	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()
	for _, conn := range s.clients {
		_ = conn.WriteJSON(status)
	}
}

// SendFiles 顺序发送一组文件
func (s *Service) SendFiles(address string, files []string) error {
	for _, file := range files {
		if err := s.sendFile(address, defaultSenderName(), file); err != nil {
			return err
		}
	}
	return nil
}

// SendFilesAs 与 SendFiles 一致，但显式带上发送方名称（用于前端展示）
func (s *Service) SendFilesAs(address, fromName string, files []string) error {
	for _, file := range files {
		if err := s.sendFile(address, fromName, file); err != nil {
			return err
		}
	}
	return nil
}

// requestUpload 在上传前先发起 /transfer/request，由对端用户批准后返回 token
func (s *Service) requestUpload(address, fromName, name string, size int64) (string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"from_name":  fromName,
		"file_name":  name,
		"total_size": size,
	})
	u := fmt.Sprintf("http://%s:%d/transfer/request", address, s.port)
	req, err := http.NewRequest("POST", u, strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 75 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request denied: %d", resp.StatusCode)
	}
	var r struct {
		Approved bool   `json:"approved"`
		Token    string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	if !r.Approved || r.Token == "" {
		return "", fmt.Errorf("rejected")
	}
	return r.Token, nil
}

// localDeviceName 返回本机用作发送者标识的名字（mDNS 注册名）
// 由调用方在 SendFiles/SendFolder 时通过参数传入；这里作为 helper 占位
func defaultSenderName() string {
	host, _ := os.Hostname()
	if host == "" {
		return "LanGive"
	}
	return host
}
// 必须带 size + tid，避免同名文件被误判为已完成
func (s *Service) queryResume(address, name string, size int64, tid string) (int64, error) {
	u := fmt.Sprintf("http://%s:%d/resume?name=%s&size=%d&tid=%s",
		address, s.port, url.QueryEscape(name), size, url.QueryEscape(tid))
	resp, err := http.Get(u)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, nil
	}
	var r struct {
		Offset    int64 `json:"offset"`
		Completed bool  `json:"completed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, err
	}
	if r.Completed {
		return -1, nil
	}
	return r.Offset, nil
}

func (s *Service) sendFile(address, fromName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	name := filepath.Base(filePath)
	transferID := uuid.New().String()
	status := &TransferStatus{
		ID:        transferID,
		Type:      "send",
		FileName:  name,
		TotalSize: stat.Size(),
		Status:    "transferring",
		PeerAddr:  address,
	}
	s.transfersMux.Lock()
	s.transfers[transferID] = status
	s.transfersMux.Unlock()
	s.broadcastProgress(status)

	// A9：先发起请求让对端用户确认，拿到一次性 token 才能 /upload
	token, err := s.requestUpload(address, fromName, name, stat.Size())
	if err != nil {
		status.Status = "failed"
		status.Error = err.Error()
		s.broadcastProgress(status)
		return err
	}

	// 询问对端续传偏移
	offset, _ := s.queryResume(address, name, stat.Size(), transferID)
	if offset == -1 {
		// 对端已经有完整文件
		status.Status = "completed"
		status.SentSize = stat.Size()
		status.Progress = 100
		s.broadcastProgress(status)
		return nil
	}
	if offset > stat.Size() {
		offset = 0
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		status.Status = "failed"
		status.Error = err.Error()
		s.broadcastProgress(status)
		return err
	}
	status.SentSize = offset
	if stat.Size() > 0 {
		status.Progress = float64(offset) / float64(stat.Size()) * 100
	}
	s.broadcastProgress(status)

	// 用 io.Pipe 包装文件流，便于在 Read 钩子里更新进度并响应取消
	ctx, cancel := context.WithCancel(s.ctx)
	s.cancelsMux.Lock()
	s.cancels[transferID] = cancel
	s.cancelsMux.Unlock()
	defer func() {
		s.cancelsMux.Lock()
		delete(s.cancels, transferID)
		s.cancelsMux.Unlock()
		cancel()
	}()

	pr, pw := io.Pipe()
	go func() {
		buf := make([]byte, chunkSize)
		sent := offset
		for {
			if ctx.Err() != nil {
				pw.CloseWithError(ctx.Err())
				return
			}
			nr, rerr := file.Read(buf)
			if nr > 0 {
				if _, werr := pw.Write(buf[:nr]); werr != nil {
					pw.CloseWithError(werr)
					return
				}
				sent += int64(nr)
				status.SentSize = sent
				if stat.Size() > 0 {
					status.Progress = float64(sent) / float64(stat.Size()) * 100
				}
				s.broadcastProgress(status)
			}
			if rerr == io.EOF {
				pw.Close()
				return
			}
			if rerr != nil {
				pw.CloseWithError(rerr)
				return
			}
		}
	}()

	url := fmt.Sprintf("http://%s:%d/upload", address, s.port)
	req, err := http.NewRequestWithContext(ctx, "POST", url, pr)
	if err != nil {
		status.Status = "failed"
		status.Error = err.Error()
		s.broadcastProgress(status)
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set(HeaderFileName, name)
	req.Header.Set(HeaderTotalSize, strconv.FormatInt(stat.Size(), 10))
	req.Header.Set(HeaderOffset, strconv.FormatInt(offset, 10))
	req.Header.Set(HeaderTransferID, transferID)
	req.Header.Set(HeaderToken, token)
	req.ContentLength = stat.Size() - offset

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			status.Status = "cancelled"
		} else {
			status.Status = "failed"
			status.Error = err.Error()
		}
		s.broadcastProgress(status)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status.Status = "failed"
		status.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		s.broadcastProgress(status)
		return fmt.Errorf("upload failed: %d", resp.StatusCode)
	}

	status.Status = "completed"
	status.Progress = 100
	status.SentSize = stat.Size()
	s.broadcastProgress(status)
	return nil
}

// SendFolder 将文件夹打包为 zip 后发送
func (s *Service) SendFolder(address string, folderPath string) error {
	return s.SendFolderAs(address, defaultSenderName(), folderPath)
}

// SendFolderAs 同 SendFolder，但带显式发送方名称
func (s *Service) SendFolderAs(address, fromName, folderPath string) error {
	tmpFile, err := os.CreateTemp("", "langive-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	zipWriter := zip.NewWriter(tmpFile)
	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(folderPath, path)
		zf, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(zf, f)
		return err
	})
	if err != nil {
		zipWriter.Close()
		tmpFile.Close()
		return err
	}
	if err := zipWriter.Close(); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// 用文件夹名 + .zip 作为目标文件名，提高对端识别度
	target := filepath.Base(folderPath) + ".zip"
	renamed := filepath.Join(filepath.Dir(tmpPath), target)
	if err := os.Rename(tmpPath, renamed); err != nil {
		return err
	}
	defer os.Remove(renamed)
	return s.sendFile(address, fromName, renamed)
}

func (s *Service) GetTransfers() []*TransferStatus {
	s.transfersMux.RLock()
	defer s.transfersMux.RUnlock()
	out := make([]*TransferStatus, 0, len(s.transfers))
	for _, t := range s.transfers {
		out = append(out, t)
	}
	return out
}

// CancelTransfer 取消传输：发送中调用 ctx cancel，已完成则返回错误
func (s *Service) CancelTransfer(id string) error {
	s.transfersMux.Lock()
	t, ok := s.transfers[id]
	if !ok {
		s.transfersMux.Unlock()
		return fmt.Errorf("transfer not found")
	}
	t.Status = "cancelled"
	s.transfersMux.Unlock()

	s.cancelsMux.Lock()
	if cancel, ok := s.cancels[id]; ok {
		cancel()
		delete(s.cancels, id)
	}
	s.cancelsMux.Unlock()
	s.broadcastProgress(t)
	return nil
}

// ClearCompleted 清除 completed/failed/cancelled 状态的传输记录
func (s *Service) ClearCompleted() {
	s.transfersMux.Lock()
	defer s.transfersMux.Unlock()
	for id, t := range s.transfers {
		switch t.Status {
		case "completed", "failed", "cancelled":
			delete(s.transfers, id)
		}
	}
}

// uniquePath 若文件已存在，自动返回 "<base> (1).<ext>" / "(2)" 等不冲突的路径
func uniquePath(p string) string {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return p
	}
	dir := filepath.Dir(p)
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	for i := 1; ; i++ {
		cand := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, i, ext))
		if _, err := os.Stat(cand); os.IsNotExist(err) {
			return cand
		}
	}
}
