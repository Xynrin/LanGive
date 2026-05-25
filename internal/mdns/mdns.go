package mdns

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
)

const (
	ServiceName = "_langive._tcp"
	Domain      = "local"
)

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Port      int    `json:"port"`
	Platform  string `json:"platform"`
	UUID      string `json:"uuid"`
	Version   string `json:"version"`
	SessionID string `json:"session_id"`
	IsPublic  bool   `json:"is_public"`
	Privacy   bool   `json:"privacy"`
	LastSeen  int64  `json:"last_seen"`
}

// Service mDNS服务
type Service struct {
	deviceName string
	deviceUUID string
	port       int
	version    string
	sessionID  string
	privacy    bool
	platform   string

	server *mdns.Server

	devices    map[string]*DeviceInfo
	devicesMux sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	onDeviceFound func(*DeviceInfo)
	onDeviceLost  func(string)
}

// NewService 创建mDNS服务
func NewService(deviceName, deviceUUID string, port int, version string, sessionID string, privacy bool) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		deviceName: deviceName,
		deviceUUID: deviceUUID,
		port:       port,
		version:    version,
		sessionID:  sessionID,
		privacy:    privacy,
		platform:   runtime.GOOS,
		devices:    make(map[string]*DeviceInfo),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// SetCallbacks 设置设备发现回调
func (s *Service) SetCallbacks(onFound func(*DeviceInfo), onLost func(string)) {
	s.onDeviceFound = onFound
	s.onDeviceLost = onLost
}

// SetDeviceName 设置设备名称
func (s *Service) SetDeviceName(name string) {
	s.deviceName = name
	s.restart()
}

// SetPrivacy 设置隐私模式
func (s *Service) SetPrivacy(enabled bool) {
	s.privacy = enabled
	s.restart()
}

// SetSession 设置会话ID
func (s *Service) SetSession(sessionID string) {
	s.sessionID = sessionID
	s.restart()
}

// restart 重启服务
func (s *Service) restart() {
	s.Stop()
	s.ctx, s.cancel = context.WithCancel(context.Background())
	_ = s.Start()
}

// Start 启动服务
func (s *Service) Start() error {
	info := s.getTXTRecords()
	service, err := mdns.NewMDNSService(
		s.deviceName,
		ServiceName,
		"",
		"",
		s.port,
		nil,
		info,
	)
	if err != nil {
		return fmt.Errorf("failed to create mDNS service: %w", err)
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return fmt.Errorf("failed to start mDNS server: %w", err)
	}
	s.server = server

	go s.discoverDevices()
	return nil
}

// Stop 停止服务
func (s *Service) Stop() {
	s.cancel()
	if s.server != nil {
		s.server.Shutdown()
		s.server = nil
	}
}

// getTXTRecords 获取TXT记录
func (s *Service) getTXTRecords() []string {
	privacyStr := "0"
	if s.privacy {
		privacyStr = "1"
	}
	return []string{
		fmt.Sprintf("name=%s", s.deviceName),
		fmt.Sprintf("platform=%s", s.platform),
		fmt.Sprintf("version=%s", s.version),
		fmt.Sprintf("uuid=%s", s.deviceUUID),
		fmt.Sprintf("session=%s", s.sessionID),
		fmt.Sprintf("privacy=%s", privacyStr),
	}
}

// discoverDevices 周期性发现设备
func (s *Service) discoverDevices() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	s.browse()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.browse()
		}
	}
}

// browse 单次浏览
func (s *Service) browse() {
	entries := make(chan *mdns.ServiceEntry, 32)

	done := make(chan struct{})
	go func() {
		for entry := range entries {
			s.handleEntry(entry)
		}
		close(done)
	}()

	params := mdns.DefaultParams(ServiceName)
	params.Domain = Domain
	params.Entries = entries
	params.Timeout = 3 * time.Second
	params.DisableIPv6 = true
	if err := mdns.Query(params); err != nil {
		// 网络不可用时静默丢弃，下一轮重试
		_ = err
	}
	close(entries)
	<-done
}

// handleEntry 处理服务条目
func (s *Service) handleEntry(entry *mdns.ServiceEntry) {
	// 跳过自己
	if entry.Name == s.deviceName+"."+ServiceName+"."+Domain+"." {
		return
	}

	// 解析TXT记录
	txtRecords := make(map[string]string)
	for _, txt := range entry.InfoFields {
		parts := strings.SplitN(txt, "=", 2)
		if len(parts) == 2 {
			txtRecords[parts[0]] = parts[1]
		}
	}

	addr := ""
	if entry.AddrV4 != nil {
		addr = entry.AddrV4.String()
	} else if entry.AddrV6IPAddr != nil {
		addr = entry.AddrV6IPAddr.String()
	}

	device := &DeviceInfo{
		ID:        entry.Name,
		Name:      txtRecords["name"],
		Address:   addr,
		Port:      entry.Port,
		Platform:  txtRecords["platform"],
		UUID:      txtRecords["uuid"],
		Version:   txtRecords["version"],
		SessionID: txtRecords["session"],
		Privacy:   txtRecords["privacy"] == "1",
		IsPublic:  txtRecords["session"] == "public",
		LastSeen:  time.Now().Unix(),
	}

	if device.UUID == "" {
		return
	}
	s.updateDevice(device)
}

// updateDevice 更新设备信息
func (s *Service) updateDevice(device *DeviceInfo) {
	s.devicesMux.Lock()
	defer s.devicesMux.Unlock()

	_, exists := s.devices[device.UUID]
	s.devices[device.UUID] = device
	if !exists && s.onDeviceFound != nil {
		s.onDeviceFound(device)
	}
}

// GetDiscoveredDevices 获取发现的设备
func (s *Service) GetDiscoveredDevices() []DeviceInfo {
	s.devicesMux.RLock()
	defer s.devicesMux.RUnlock()

	devices := make([]DeviceInfo, 0, len(s.devices))
	for _, d := range s.devices {
		devices = append(devices, *d)
	}
	return devices
}

// GetPublicDevices 获取公共会话的设备
func (s *Service) GetPublicDevices() []DeviceInfo {
	s.devicesMux.RLock()
	defer s.devicesMux.RUnlock()

	devices := make([]DeviceInfo, 0)
	for _, d := range s.devices {
		if d.IsPublic && !d.Privacy {
			devices = append(devices, *d)
		}
	}
	return devices
}

// GetDevice 根据UUID获取设备
func (s *Service) GetDevice(uuid string) *DeviceInfo {
	s.devicesMux.RLock()
	defer s.devicesMux.RUnlock()
	return s.devices[uuid]
}

// RemoveStaleDevices 移除超时的设备
func (s *Service) RemoveStaleDevices(timeout time.Duration) {
	s.devicesMux.Lock()
	defer s.devicesMux.Unlock()

	threshold := time.Now().Unix() - int64(timeout.Seconds())
	for uuid, device := range s.devices {
		if device.LastSeen < threshold {
			delete(s.devices, uuid)
			if s.onDeviceLost != nil {
				s.onDeviceLost(uuid)
			}
		}
	}
}

// StartCleanupRoutine 启动清理超时设备的goroutine
func (s *Service) StartCleanupRoutine(interval, timeout time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.RemoveStaleDevices(timeout)
			}
		}
	}()
}
