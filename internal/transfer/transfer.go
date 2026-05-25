package transfer

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

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

	ctx    context.Context
	cancel context.CancelFunc
}

func NewService(downloadPath string, port int) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	return &Service{
		downloadPath: downloadPath,
		port:         port,
		router:       router,
		transfers:    make(map[string]*TransferStatus),
		cancels:      make(map[string]context.CancelFunc),
		clients:      make(map[string]*websocket.Conn),
		ctx:          ctx,
		cancel:       cancel,
	}
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

// handleResume 返回服务端 <file>.part 已写入的字节数，供发送端断点续传
func (s *Service) handleResume(c *gin.Context) {
	name := filepath.Base(c.Query("name"))
	if name == "" || name == "." || name == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing name"})
		return
	}
	partPath := filepath.Join(s.downloadPath, name+partSuffix)
	finalPath := filepath.Join(s.downloadPath, name)

	if st, err := os.Stat(finalPath); err == nil {
		c.JSON(http.StatusOK, gin.H{"offset": st.Size(), "completed": true})
		return
	}
	if st, err := os.Stat(partPath); err == nil {
		c.JSON(http.StatusOK, gin.H{"offset": st.Size(), "completed": false})
		return
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

	name := filepath.Base(c.GetHeader(HeaderFileName))
	if name == "" || name == "." {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing filename header"})
		return
	}
	total, _ := strconv.ParseInt(c.GetHeader(HeaderTotalSize), 10, 64)
	offset, _ := strconv.ParseInt(c.GetHeader(HeaderOffset), 10, 64)

	partPath := filepath.Join(s.downloadPath, name+partSuffix)
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

	transferID := uuid.New().String()
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

	// 全部接收完毕：去掉 .part 后缀
	if total == 0 || written >= total {
		dst.Close()
		if err := os.Rename(partPath, finalPath); err != nil {
			status.Status = "failed"
			status.Error = err.Error()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		status.Status = "completed"
		status.Progress = 100
	} else {
		status.Status = "paused"
	}
	s.broadcastProgress(status)
	c.JSON(http.StatusOK, gin.H{"id": transferID, "received": written, "status": status.Status})
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
		if err := s.sendFile(address, file); err != nil {
			return err
		}
	}
	return nil
}

// queryResume 询问对端已收字节数
func (s *Service) queryResume(address, name string) (int64, error) {
	url := fmt.Sprintf("http://%s:%d/resume?name=%s", address, s.port, name)
	resp, err := http.Get(url)
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

func (s *Service) sendFile(address string, filePath string) error {
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

	// 询问对端续传偏移
	offset, _ := s.queryResume(address, name)
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
	return s.sendFile(address, renamed)
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
