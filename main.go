package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/Xynrin/LanGive/internal/config"
	"github.com/Xynrin/LanGive/internal/mdns"
	"github.com/Xynrin/LanGive/internal/security"
	"github.com/Xynrin/LanGive/internal/transfer"
	"github.com/Xynrin/LanGive/internal/updater"
)

//go:embed all:frontend/dist
var assets embed.FS

type LanGiveApp struct {
	ctx      context.Context
	config   *config.Config
	mdns     *mdns.Service
	transfer *transfer.Service
	updater  *updater.Service
	security *security.Manager
}

func NewLanGiveApp() *LanGiveApp {
	return &LanGiveApp{}
}

func (a *LanGiveApp) startup(ctx context.Context) {
	a.ctx = ctx

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	a.config = cfg

	// Initialize security manager
	a.security = security.NewSecurityManager()
	a.security.CreatePublicSession()

	// Initialize mDNS service
	a.mdns = mdns.NewService(
		cfg.DeviceName,
		cfg.DeviceUUID,
		cfg.Port,
		cfg.Version,
		cfg.SessionID,
		cfg.PrivacyMode,
	)
	if err := a.mdns.Start(); err != nil {
		fmt.Printf("Failed to start mDNS: %v\n", err)
	}
	a.mdns.StartCleanupRoutine(cfg.GetScanInterval(), cfg.GetDeviceTimeout())

	// Initialize transfer service
	a.transfer = transfer.NewService(cfg.DownloadPath, cfg.Port)
	a.transfer.SetOnIncomingRequest(func(r *transfer.IncomingRequest) {
		if a.ctx != nil {
			wailsruntime.EventsEmit(a.ctx, "transfer:incoming", r)
		}
	})
	if err := a.transfer.Start(); err != nil {
		fmt.Printf("Failed to start transfer service: %v\n", err)
	}

	// Initialize updater service
	a.updater = updater.NewService(cfg.Version)
}

func (a *LanGiveApp) shutdown(ctx context.Context) {
	if a.mdns != nil {
		a.mdns.Stop()
	}
	if a.transfer != nil {
		a.transfer.Stop()
	}
}

// ============ Device Methods ============

// GetDevices returns discovered devices
func (a *LanGiveApp) GetDevices() []mdns.DeviceInfo {
	return a.mdns.GetDiscoveredDevices()
}

// GetPublicDevices returns only public session devices
func (a *LanGiveApp) GetPublicDevices() []mdns.DeviceInfo {
	return a.mdns.GetPublicDevices()
}

// GetDevice returns a specific device by UUID
func (a *LanGiveApp) GetDevice(uuid string) *mdns.DeviceInfo {
	return a.mdns.GetDevice(uuid)
}

// ============ Transfer Methods ============

// SendFiles sends files to a device
func (a *LanGiveApp) SendFiles(deviceID string, files []string) error {
	device := a.mdns.GetDevice(deviceID)
	if device == nil {
		return fmt.Errorf("device not found")
	}
	return a.transfer.SendFiles(device.Address, files)
}

// SendFolder sends a folder to a device
func (a *LanGiveApp) SendFolder(deviceID string, folderPath string) error {
	device := a.mdns.GetDevice(deviceID)
	if device == nil {
		return fmt.Errorf("device not found")
	}
	return a.transfer.SendFolder(device.Address, folderPath)
}

// IsDirectory 判断给定路径是否为目录；不存在或读取失败返回 false + err
func (a *LanGiveApp) IsDirectory(path string) (bool, error) {
	st, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return st.IsDir(), nil
}

// SendPath 按 path 实际类型派发：目录 → SendFolder，文件 → SendFiles
func (a *LanGiveApp) SendPath(deviceID string, path string) error {
	device := a.mdns.GetDevice(deviceID)
	if device == nil {
		return fmt.Errorf("device not found")
	}
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return a.transfer.SendFolder(device.Address, path)
	}
	return a.transfer.SendFiles(device.Address, []string{path})
}

// GetTransfers returns all transfers
func (a *LanGiveApp) GetTransfers() []*transfer.TransferStatus {
	return a.transfer.GetTransfers()
}

// CancelTransfer cancels a transfer
func (a *LanGiveApp) CancelTransfer(id string) error {
	return a.transfer.CancelTransfer(id)
}

// ClearCompletedTransfers 清空已完成/失败/取消的传输记录
func (a *LanGiveApp) ClearCompletedTransfers() error {
	a.transfer.ClearCompleted()
	return nil
}

// PendingIncomingRequests 返回当前等待用户确认的传入请求
func (a *LanGiveApp) PendingIncomingRequests() []*transfer.IncomingRequest {
	return a.transfer.PendingRequests()
}

// ApproveIncoming 批准一个传入请求并发放 token
func (a *LanGiveApp) ApproveIncoming(id string) error {
	_, err := a.transfer.ApproveIncoming(id)
	return err
}

// RejectIncoming 拒绝一个传入请求
func (a *LanGiveApp) RejectIncoming(id string) error {
	return a.transfer.RejectIncoming(id)
}

// ============ Config Methods ============

// GetDeviceName returns device name
func (a *LanGiveApp) GetDeviceName() string {
	return a.config.DeviceName
}

// SetDeviceName sets device name
func (a *LanGiveApp) SetDeviceName(name string) error {
	a.config.DeviceName = name
	if err := a.config.Save(); err != nil {
		return err
	}
	a.mdns.SetDeviceName(name)
	return nil
}

// GetDeviceUUID returns device UUID
func (a *LanGiveApp) GetDeviceUUID() string {
	return a.config.DeviceUUID
}

// GetDownloadPath returns download path
func (a *LanGiveApp) GetDownloadPath() string {
	return a.config.DownloadPath
}

// SetDownloadPath sets download path
func (a *LanGiveApp) SetDownloadPath(path string) error {
	a.config.DownloadPath = path
	return a.config.Save()
}

// GetPort returns service port
func (a *LanGiveApp) GetPort() int {
	return a.config.Port
}

// SetPort sets service port
func (a *LanGiveApp) SetPort(port int) error {
	a.config.Port = port
	return a.config.Save()
}

// GetScanInterval returns scan interval
func (a *LanGiveApp) GetScanInterval() int {
	return a.config.ScanInterval
}

// SetScanInterval sets scan interval
func (a *LanGiveApp) SetScanInterval(interval int) error {
	a.config.ScanInterval = interval
	return a.config.Save()
}

// GetPrivacyMode returns privacy mode status
func (a *LanGiveApp) GetPrivacyMode() bool {
	return a.config.PrivacyMode
}

// SetPrivacyMode sets privacy mode
func (a *LanGiveApp) SetPrivacyMode(enabled bool) error {
	a.config.SetPrivacyMode(enabled)
	if err := a.config.Save(); err != nil {
		return err
	}
	a.mdns.SetPrivacy(enabled)
	return nil
}

// GetSessionID returns current session ID
func (a *LanGiveApp) GetSessionID() string {
	return a.config.SessionID
}

// GetAutoUpdate returns auto update status
func (a *LanGiveApp) GetAutoUpdate() bool {
	return a.config.AutoUpdate
}

// SetAutoUpdate sets auto update status
func (a *LanGiveApp) SetAutoUpdate(enabled bool) error {
	a.config.AutoUpdate = enabled
	return a.config.Save()
}

// ResetConfig resets all configuration to default
func (a *LanGiveApp) ResetConfig() error {
	newConfig := &config.Config{
		DeviceName:   config.GetDefaultDeviceName(),
		DeviceUUID:   a.config.DeviceUUID, // Keep UUID
		DownloadPath: config.GetDefaultDownloadPath(),
		Port:         config.DefaultPort,
		PrivacyMode:  false,
		SessionID:    config.PublicSession,
		Version:      config.Version,
		AutoUpdate:   true,
		ScanInterval: int(config.ForegroundScanInterval.Seconds()),
	}
	*a.config = *newConfig
	return a.config.Save()
}

// ============ Dialog Methods ============

// SelectFolder opens folder dialog
func (a *LanGiveApp) SelectFolder() (string, error) {
	return wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:                "选择文件夹",
		CanCreateDirectories: true,
	})
}

// SelectFiles opens file dialog
func (a *LanGiveApp) SelectFiles() ([]string, error) {
	return wailsruntime.OpenMultipleFilesDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title:                "选择文件",
		CanCreateDirectories: false,
	})
}

// OpenInExplorer 用系统资源管理器打开指定路径
func (a *LanGiveApp) OpenInExplorer(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

// ============ Update Methods ============

// CheckUpdate checks for updates
func (a *LanGiveApp) CheckUpdate() (*updater.UpdateInfo, error) {
	return a.updater.CheckUpdate()
}

// GetVersion returns app version
func (a *LanGiveApp) GetVersion() string {
	return a.config.Version
}

// DownloadAndInstall downloads and installs update
func (a *LanGiveApp) DownloadAndInstall(url string) error {
	return a.updater.DownloadAndInstall(url)
}

// Restart restarts the app
func (a *LanGiveApp) Restart() error {
	return a.updater.Restart()
}

func main() {
	app := NewLanGiveApp()

	err := wails.Run(&options.App{
		Title:     "LanGive",
		Width:     1200,
		Height:    800,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 250, B: 252, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
