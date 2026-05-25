// Package config 是 internal/config 的 gomobile 友好包装。
// gomobile bind 不允许绑定 internal/* 包，因此这里做一层薄壳转发。
package config

import (
	icfg "github.com/Xynrin/LanGive/internal/config"
)

// Config 暴露给 Java 的配置对象（gomobile 会为每个字段生成 getter/setter）。
type Config struct {
	inner *icfg.Config

	DeviceName     string
	DeviceUUID     string
	DeviceToken    string
	DownloadPath   string
	Port           int64
	PrivacyMode    bool
	SessionID      string
	Version        string
	AutoUpdate     bool
	ScanInterval   int64
	FirstRun       bool
	BackgroundMode bool
}

// Load 加载或创建配置。
func Load() (*Config, error) {
	c, err := icfg.Load()
	if err != nil {
		return nil, err
	}
	return wrap(c), nil
}

func wrap(c *icfg.Config) *Config {
	return &Config{
		inner:          c,
		DeviceName:     c.DeviceName,
		DeviceUUID:     c.DeviceUUID,
		DeviceToken:    c.DeviceToken,
		DownloadPath:   c.DownloadPath,
		Port:           int64(c.Port),
		PrivacyMode:    c.PrivacyMode,
		SessionID:      c.SessionID,
		Version:        c.Version,
		AutoUpdate:     c.AutoUpdate,
		ScanInterval:   int64(c.ScanInterval),
		FirstRun:       c.FirstRun,
		BackgroundMode: c.BackgroundMode,
	}
}

func (c *Config) sync() {
	c.inner.DeviceName = c.DeviceName
	c.inner.DeviceUUID = c.DeviceUUID
	c.inner.DeviceToken = c.DeviceToken
	c.inner.DownloadPath = c.DownloadPath
	c.inner.Port = int(c.Port)
	c.inner.PrivacyMode = c.PrivacyMode
	c.inner.SessionID = c.SessionID
	c.inner.Version = c.Version
	c.inner.AutoUpdate = c.AutoUpdate
	c.inner.ScanInterval = int(c.ScanInterval)
	c.inner.FirstRun = c.FirstRun
	c.inner.BackgroundMode = c.BackgroundMode
}

// Save 持久化配置。
func (c *Config) Save() error {
	c.sync()
	return c.inner.Save()
}

// GetScanInterval 返回当前扫描间隔（纳秒）。
func (c *Config) GetScanInterval() int64 {
	c.sync()
	return int64(c.inner.GetScanInterval())
}

// GetDeviceTimeout 返回设备超时（纳秒），约定 = ScanInterval × 3。
func (c *Config) GetDeviceTimeout() int64 {
	c.sync()
	return int64(c.inner.GetDeviceTimeout())
}

// SetBackgroundMode 切换后台模式（自动调整扫描间隔）。
func (c *Config) SetBackgroundMode(background bool) {
	c.sync()
	c.inner.SetBackgroundMode(background)
	c.BackgroundMode = c.inner.BackgroundMode
	c.ScanInterval = int64(c.inner.ScanInterval)
}

// SetPrivacyMode 切换隐私模式（生成新 SessionID）。
func (c *Config) SetPrivacyMode(enabled bool) {
	c.sync()
	c.inner.SetPrivacyMode(enabled)
	c.PrivacyMode = c.inner.PrivacyMode
	c.SessionID = c.inner.SessionID
}

// GetDefaultDownloadPath 默认下载目录。
func GetDefaultDownloadPath() string { return icfg.GetDefaultDownloadPath() }

// GetDefaultDeviceName 默认设备名（系统主机名）。
func GetDefaultDeviceName() string { return icfg.GetDefaultDeviceName() }

// Version 当前 LanGive 版本号。
func Version() string { return icfg.Version }
