package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/google/uuid"
)

const (
	Version        = "1.0.13"
	DefaultPort    = 5566
	PublicSession  = "public"

	// 扫描间隔（秒）
	ForegroundScanInterval = 5 * time.Second
	BackgroundScanInterval = 30 * time.Second
	DeviceTimeout         = 15 * time.Second
)

type Config struct {
	// 设备配置
	DeviceName   string `json:"device_name"`   // 设备显示名称
	DeviceUUID   string `json:"device_uuid"`   // 设备唯一标识
	DeviceToken  string `json:"device_token"`  // 设备连接令牌

	// 存储配置
	DownloadPath string `json:"download_path"` // 下载目录
	Port         int    `json:"port"`          // 服务端口

	// 会话与隐私
	PrivacyMode  bool   `json:"privacy_mode"`  // 隐私模式
	SessionID    string `json:"session_id"`    // 会话ID (public 或自定义)

	// 应用配置
	Version      string `json:"version"`       // 当前版本
	AutoUpdate   bool   `json:"auto_update"`   // 自动更新
	ScanInterval int    `json:"scan_interval"` // 扫描间隔（秒）

	// 运行状态
	FirstRun     bool   `json:"first_run"`     // 是否首次运行
	BackgroundMode bool `json:"background_mode"` // 是否后台运行
}

func getConfigDir() string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Local", "LanGive")
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "LanGive")
	default: // linux and others
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(home, ".config")
		}
		return filepath.Join(configDir, "langive")
	}
}

// GetDefaultDownloadPath 返回默认下载目录
func GetDefaultDownloadPath() string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "Downloads", "LanGive")
	case "darwin":
		return filepath.Join(home, "Downloads", "LanGive")
	default:
		return filepath.Join(home, "Downloads", "LanGive")
	}
}

// GetDefaultDeviceName 返回默认设备名（系统主机名）
func GetDefaultDeviceName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "LanGive Device"
	}
	return hostname
}

func generateToken() string {
	return uuid.New().String()
}

func Load() (*Config, error) {
	configDir := getConfigDir()
	configPath := filepath.Join(configDir, "config.json")

	// 尝试读取现有配置
	data, err := os.ReadFile(configPath)
	if err == nil {
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err == nil {
			// 确保版本号是最新的
			cfg.Version = Version
			cfg.BackgroundMode = false
			return &cfg, nil
		}
	}

	// 创建默认配置
	cfg := &Config{
		DeviceName:   GetDefaultDeviceName(),
		DeviceUUID:   uuid.New().String(),
		DeviceToken:  generateToken(),
		DownloadPath: GetDefaultDownloadPath(),
		Port:         DefaultPort,
		PrivacyMode:  false,
		SessionID:    PublicSession,
		Version:      Version,
		AutoUpdate:   true,
		ScanInterval: int(ForegroundScanInterval.Seconds()),
		FirstRun:     true,
		BackgroundMode: false,
	}

	// 确保下载目录存在
	os.MkdirAll(cfg.DownloadPath, 0755)

	// 保存配置
	if err := cfg.Save(); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Save() error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetScanInterval 根据运行模式返回扫描间隔
func (c *Config) GetScanInterval() time.Duration {
	if c.BackgroundMode {
		return BackgroundScanInterval
	}
	return time.Duration(c.ScanInterval) * time.Second
}

// GetDeviceTimeout 设备超时时间，约定为扫描间隔的 3 倍
// 避免后台模式下扫描慢于固定 timeout 导致设备列表闪烁
func (c *Config) GetDeviceTimeout() time.Duration {
	return c.GetScanInterval() * 3
}

// SetBackgroundMode 设置后台模式并调整扫描间隔
func (c *Config) SetBackgroundMode(background bool) {
	c.BackgroundMode = background
	if background && c.ScanInterval == int(ForegroundScanInterval.Seconds()) {
		c.ScanInterval = int(BackgroundScanInterval.Seconds())
	}
}

// SetPrivacyMode 切换隐私模式
func (c *Config) SetPrivacyMode(enabled bool) {
	c.PrivacyMode = enabled
	if enabled {
		c.SessionID = uuid.New().String() // 为隐私模式生成独立会话ID
	} else {
		c.SessionID = PublicSession
	}
}
