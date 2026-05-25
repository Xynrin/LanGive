// Package transfer 是 internal/transfer 的 gomobile 友好包装。
package transfer

import (
	itransfer "github.com/Xynrin/LanGive/internal/transfer"
)

// IncomingRequest 暴露给 Java 的待确认请求。
type IncomingRequest struct {
	ID         string
	FromName   string
	FromAddr   string
	FileName   string
	TotalSize  int64
	ReceivedAt int64
}

func wrapRequest(r *itransfer.IncomingRequest) *IncomingRequest {
	if r == nil {
		return nil
	}
	return &IncomingRequest{
		ID:         r.ID,
		FromName:   r.FromName,
		FromAddr:   r.FromAddr,
		FileName:   r.FileName,
		TotalSize:  r.TotalSize,
		ReceivedAt: r.ReceivedAt,
	}
}

// IncomingRequestSlice gomobile 不支持返回 []*Struct，包装一层。
type IncomingRequestSlice struct {
	items []*IncomingRequest
}

func (s *IncomingRequestSlice) Size() int64 { return int64(len(s.items)) }
func (s *IncomingRequestSlice) Get(i int64) *IncomingRequest {
	if i < 0 || i >= int64(len(s.items)) {
		return nil
	}
	return s.items[i]
}

// TransferStatus 暴露给 Java 的传输状态。
type TransferStatus struct {
	ID        string
	Type      string
	FileName  string
	TotalSize int64
	SentSize  int64
	Progress  float64
	Status    string
	Error     string
	PeerAddr  string
}

func wrapStatus(s *itransfer.TransferStatus) *TransferStatus {
	if s == nil {
		return nil
	}
	return &TransferStatus{
		ID:        s.ID,
		Type:      s.Type,
		FileName:  s.FileName,
		TotalSize: s.TotalSize,
		SentSize:  s.SentSize,
		Progress:  s.Progress,
		Status:    s.Status,
		Error:     s.Error,
		PeerAddr:  s.PeerAddr,
	}
}

// TransferStatusSlice 包装。
type TransferStatusSlice struct {
	items []*TransferStatus
}

func (s *TransferStatusSlice) Size() int64 { return int64(len(s.items)) }
func (s *TransferStatusSlice) Get(i int64) *TransferStatus {
	if i < 0 || i >= int64(len(s.items)) {
		return nil
	}
	return s.items[i]
}

// StringSlice gomobile 不支持 []string 入参，包装一层。
type StringSlice struct {
	items []string
}

func NewStringSlice() *StringSlice            { return &StringSlice{} }
func (s *StringSlice) Add(v string)           { s.items = append(s.items, v) }
func (s *StringSlice) Size() int64            { return int64(len(s.items)) }
func (s *StringSlice) Get(i int64) string {
	if i < 0 || i >= int64(len(s.items)) {
		return ""
	}
	return s.items[i]
}

// IncomingRequestHandler Java 实现该接口，Service 在收到请求时回调。
type IncomingRequestHandler interface {
	OnIncomingRequest(req *IncomingRequest)
}

// Service 包装 internal/transfer.Service。
type Service struct {
	inner *itransfer.Service
}

// NewService Java 侧通过 transfer.Transfer.newService 调用。
func NewService(downloadPath string, port int64) *Service {
	return &Service{inner: itransfer.NewService(downloadPath, int(port))}
}

func (s *Service) Start() error { return s.inner.Start() }
func (s *Service) Stop()        { s.inner.Stop() }

func (s *Service) SetOnIncomingRequest(handler IncomingRequestHandler) {
	if handler == nil {
		s.inner.SetOnIncomingRequest(nil)
		return
	}
	s.inner.SetOnIncomingRequest(func(r *itransfer.IncomingRequest) {
		handler.OnIncomingRequest(wrapRequest(r))
	})
}

func (s *Service) ApproveIncoming(id string) (string, error) {
	return s.inner.ApproveIncoming(id)
}
func (s *Service) RejectIncoming(id string) error { return s.inner.RejectIncoming(id) }

func (s *Service) PendingRequests() *IncomingRequestSlice {
	src := s.inner.PendingRequests()
	out := make([]*IncomingRequest, 0, len(src))
	for _, r := range src {
		out = append(out, wrapRequest(r))
	}
	return &IncomingRequestSlice{items: out}
}

func (s *Service) GetTransfers() *TransferStatusSlice {
	src := s.inner.GetTransfers()
	out := make([]*TransferStatus, 0, len(src))
	for _, t := range src {
		out = append(out, wrapStatus(t))
	}
	return &TransferStatusSlice{items: out}
}

func (s *Service) CancelTransfer(id string) error { return s.inner.CancelTransfer(id) }
func (s *Service) ClearCompleted()                { s.inner.ClearCompleted() }

func (s *Service) SendFilesAs(address, fromName string, files *StringSlice) error {
	if files == nil {
		return nil
	}
	return s.inner.SendFilesAs(address, fromName, files.items)
}

func (s *Service) SendFolderAs(address, fromName, folderPath string) error {
	return s.inner.SendFolderAs(address, fromName, folderPath)
}

// NewService 备用别名（Java 侧 Transfer.newService 直接对应这里）。
// 注：gomobile 会把包级函数 NewService 暴露为 transfer.NewService。
