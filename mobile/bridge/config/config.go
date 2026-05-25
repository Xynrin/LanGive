// Package config 是 internal/config 的 gomobile 友好包装。
// 字段全部不导出，避免 gomobile 自动生成的 setX/getX 与显式方法重定义冲突。
package config

import (
	icfg "github.com/Xynrin/LanGive/internal/config"
)

// Config 暴露给 Java 的配置对象。
type Config struct {
	inner *icfg.Config
}

// Load 加载或创建配置。
func Load() (*Config, error) {
	c, err := icfg.Load()
	if err != nil {
		return nil, err
	}
	return &Config{inner: c}, nil
}

// 字段读取
func (c *Config) GetDeviceName() string     { return c.inner.DeviceName }
func (c *Config) GetDeviceUUID() string     { return c.inner.DeviceUUID }
func (c *Config) GetDeviceToken() string    { return c.inner.DeviceToken }
func (c *Config) GetDownloadPath() string   { return c.inner.DownloadPath }
func (c *Config) GetPort() int64            { return int64(c.inner.Port) }
func (c *Config) GetPrivacyMode() bool      { return c.inner.PrivacyMode }
func (c *Config) GetSessionID() string      { return c.inner.SessionID }
func (c *Config) GetVersion() string        { return c.inner.Version }
func (c *Config) GetAutoUpdate() bool       { return c.inner.AutoUpdate }
func (c *Config) GetFirstRun() bool         { return c.inner.FirstRun }
func (c *Config) GetBackgroundMode() bool   { return c.inner.BackgroundMode }

// 字段写入
func (c *Config) SetDeviceName(v string)   { c.inner.DeviceName = v }
func (c *Config) SetDeviceUUID(v string)   { c.inner.DeviceUUID = v }
func (c *Config) SetDeviceToken(v string)  { c.inner.DeviceToken = v }
func (c *Config) SetDownloadPath(v string) { c.inner.DownloadPath = v }
func (c *Config) SetPort(v int64)          { c.inner.Port = int(v) }
func (c *Config) SetSessionID(v string)    { c.inner.SessionID = v }
func (c *Config) SetVersion(v string)      { c.inner.Version = v }
func (c *Config) SetAutoUpdate(v bool)     { c.inner.AutoUpdate = v }
func (c *Config) SetFirstRun(v bool)       { c.inner.FirstRun = v }

// Save 持久化配置。
func (c *Config) Save() error { return c.inner.Save() }

// GetScanInterval 当前扫描间隔（纳秒，自动按前台/后台模式切换）。
func (c *Config) GetScanInterval() int64 { return int64(c.inner.GetScanInterval()) }

// GetScanIntervalSeconds 配置中存储的扫描间隔（秒）。
func (c *Config) GetScanIntervalSeconds() int64 { return int64(c.inner.ScanInterval) }

// SetScanIntervalSeconds 设置扫描间隔（秒）。
func (c *Config) SetScanIntervalSeconds(s int64) { c.inner.ScanInterval = int(s) }

// GetDeviceTimeout 设备超时（纳秒），= ScanInterval × 3。
func (c *Config) GetDeviceTimeout() int64 { return int64(c.inner.GetDeviceTimeout()) }

// SetBackgroundMode 切换后台模式（自动调整扫描间隔）。
func (c *Config) SetBackgroundMode(background bool) { c.inner.SetBackgroundMode(background) }

// SetPrivacyMode 切换隐私模式（生成新 SessionID）。
func (c *Config) SetPrivacyMode(enabled bool) { c.inner.SetPrivacyMode(enabled) }

// GetDefaultDownloadPath 默认下载目录。
func GetDefaultDownloadPath() string { return icfg.GetDefaultDownloadPath() }

// GetDefaultDeviceName 默认设备名（系统主机名）。
func GetDefaultDeviceName() string { return icfg.GetDefaultDeviceName() }

// CurrentVersion 当前 LanGive 版本号。
func CurrentVersion() string { return icfg.Version }
