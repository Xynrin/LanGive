// Package mdns 是 internal/mdns 的 gomobile 友好包装。
package mdns

import (
	"time"

	imdns "github.com/Xynrin/LanGive/internal/mdns"
)

// DeviceInfo 暴露给 Java 的设备信息。
type DeviceInfo struct {
	ID        string
	Name      string
	Address   string
	Port      int64
	Platform  string
	UUID      string
	Version   string
	SessionID string
	IsPublic  bool
	Privacy   bool
	LastSeen  int64
}

func wrapDevice(d imdns.DeviceInfo) *DeviceInfo {
	return &DeviceInfo{
		ID:        d.ID,
		Name:      d.Name,
		Address:   d.Address,
		Port:      int64(d.Port),
		Platform:  d.Platform,
		UUID:      d.UUID,
		Version:   d.Version,
		SessionID: d.SessionID,
		IsPublic:  d.IsPublic,
		Privacy:   d.Privacy,
		LastSeen:  d.LastSeen,
	}
}

// DeviceInfoSlice gomobile 不支持 []*Struct，包装一层。
type DeviceInfoSlice struct {
	items []*DeviceInfo
}

func (s *DeviceInfoSlice) Size() int64 { return int64(len(s.items)) }
func (s *DeviceInfoSlice) Get(i int64) *DeviceInfo {
	if i < 0 || i >= int64(len(s.items)) {
		return nil
	}
	return s.items[i]
}

// Service 包装 internal/mdns.Service。
type Service struct {
	inner *imdns.Service
}

// NewService 与 internal API 同形参，方便 Java 侧迁移。
func NewService(deviceName, deviceUUID string, port int64, version, sessionID string, privacy bool) *Service {
	return &Service{inner: imdns.NewService(deviceName, deviceUUID, int(port), version, sessionID, privacy)}
}

func (s *Service) Start() error             { return s.inner.Start() }
func (s *Service) Stop()                    { s.inner.Stop() }
func (s *Service) SetDeviceName(name string) { s.inner.SetDeviceName(name) }
func (s *Service) SetPrivacy(enabled bool)   { s.inner.SetPrivacy(enabled) }
func (s *Service) SetSession(sessionID string) { s.inner.SetSession(sessionID) }

func (s *Service) GetDiscoveredDevices() *DeviceInfoSlice {
	src := s.inner.GetDiscoveredDevices()
	out := make([]*DeviceInfo, 0, len(src))
	for _, d := range src {
		out = append(out, wrapDevice(d))
	}
	return &DeviceInfoSlice{items: out}
}

func (s *Service) GetPublicDevices() *DeviceInfoSlice {
	src := s.inner.GetPublicDevices()
	out := make([]*DeviceInfo, 0, len(src))
	for _, d := range src {
		out = append(out, wrapDevice(d))
	}
	return &DeviceInfoSlice{items: out}
}

func (s *Service) GetDevice(uuid string) *DeviceInfo {
	d := s.inner.GetDevice(uuid)
	if d == nil {
		return nil
	}
	return wrapDevice(*d)
}

// StartCleanupRoutine interval / timeout 单位为纳秒（与 GetScanInterval 对齐）。
func (s *Service) StartCleanupRoutine(intervalNs, timeoutNs int64) {
	s.inner.StartCleanupRoutine(time.Duration(intervalNs), time.Duration(timeoutNs))
}

func (s *Service) RemoveStaleDevices(timeoutNs int64) {
	s.inner.RemoveStaleDevices(time.Duration(timeoutNs))
}
