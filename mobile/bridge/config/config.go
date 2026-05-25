// Package config 是 internal/config 的 gomobile 友好包装。
// 结构体名故意叫 Settings，避免和 gomobile 自动生成的包级类 Config 撞名。
package config

import (
	icfg "github.com/Xynrin/LanGive/internal/config"
)

// Settings 暴露给 Java 的配置对象。
type Settings struct {
	inner *icfg.Config
}

// Load 加载或创建配置。
func Load() (*Settings, error) {
	c, err := icfg.Load()
	if err != nil {
		return nil, err
	}
	return &Settings{inner: c}, nil
}

func (s *Settings) GetDeviceName() string     { return s.inner.DeviceName }
func (s *Settings) GetDeviceUUID() string     { return s.inner.DeviceUUID }
func (s *Settings) GetDeviceToken() string    { return s.inner.DeviceToken }
func (s *Settings) GetDownloadPath() string   { return s.inner.DownloadPath }
func (s *Settings) GetPort() int64            { return int64(s.inner.Port) }
func (s *Settings) GetPrivacyMode() bool      { return s.inner.PrivacyMode }
func (s *Settings) GetSessionID() string      { return s.inner.SessionID }
func (s *Settings) GetVersion() string        { return s.inner.Version }
func (s *Settings) GetAutoUpdate() bool       { return s.inner.AutoUpdate }
func (s *Settings) GetFirstRun() bool         { return s.inner.FirstRun }
func (s *Settings) GetBackgroundMode() bool   { return s.inner.BackgroundMode }

// GetScanInterval 配置中存储的扫描间隔（秒）。
func (s *Settings) GetScanInterval() int64 { return int64(s.inner.ScanInterval) }

func (s *Settings) SetDeviceName(v string)   { s.inner.DeviceName = v }
func (s *Settings) SetDeviceUUID(v string)   { s.inner.DeviceUUID = v }
func (s *Settings) SetDeviceToken(v string)  { s.inner.DeviceToken = v }
func (s *Settings) SetDownloadPath(v string) { s.inner.DownloadPath = v }
func (s *Settings) SetPort(v int64)          { s.inner.Port = int(v) }
func (s *Settings) SetSessionID(v string)    { s.inner.SessionID = v }
func (s *Settings) SetVersion(v string)      { s.inner.Version = v }
func (s *Settings) SetAutoUpdate(v bool)     { s.inner.AutoUpdate = v }
func (s *Settings) SetFirstRun(v bool)       { s.inner.FirstRun = v }
func (s *Settings) SetScanInterval(v int64)  { s.inner.ScanInterval = int(v) }

// Save 持久化配置。
func (s *Settings) Save() error { return s.inner.Save() }

// GetDeviceTimeout 设备超时（秒），= 当前扫描间隔（含前后台模式调整）× 3。
func (s *Settings) GetDeviceTimeout() int64 {
	return int64(s.inner.GetDeviceTimeout() / 1e9)
}

// SetBackgroundMode 切换后台模式（自动调整扫描间隔）。
func (s *Settings) SetBackgroundMode(background bool) { s.inner.SetBackgroundMode(background) }

// SetPrivacyMode 切换隐私模式（生成新 SessionID）。
func (s *Settings) SetPrivacyMode(enabled bool) { s.inner.SetPrivacyMode(enabled) }

// GetDefaultDownloadPath 默认下载目录。
func GetDefaultDownloadPath() string { return icfg.GetDefaultDownloadPath() }

// GetDefaultDeviceName 默认设备名（系统主机名）。
func GetDefaultDeviceName() string { return icfg.GetDefaultDeviceName() }

// CurrentVersion 当前 LanGive 版本号。
func CurrentVersion() string { return icfg.Version }
