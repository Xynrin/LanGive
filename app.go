package main

import (
	"context"
	"fmt"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetSystemInfo returns system information
func (a *App) GetSystemInfo() map[string]string {
	return map[string]string{
		"os":   goruntime.GOOS,
		"arch": goruntime.GOARCH,
	}
}

// SelectFolder opens a folder dialog and returns the selected path
func (a *App) SelectFolder() (string, error) {
	options := runtime.OpenDialogOptions{
		Title: "选择文件夹",
		CanCreateDirectories: true,
	}
	return runtime.OpenDirectoryDialog(a.ctx, options)
}

// SelectFiles opens a file dialog and returns selected file paths
func (a *App) SelectFiles() ([]string, error) {
	options := runtime.OpenDialogOptions{
		Title: "选择文件",
		CanCreateDirectories: false,
	}
	return runtime.OpenMultipleFilesDialog(a.ctx, options)
}

// ShowError shows an error dialog
func (a *App) ShowError(title, message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   title,
		Message: message,
	})
}

// ShowInfo shows an info dialog
func (a *App) ShowInfo(title, message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
}

// GetAppMenu returns the application menu
func GetAppMenu(app *App) *menu.Menu {
	AppMenu := menu.NewMenu()

	// File menu
	FileMenu := AppMenu.AddSubmenu("文件")
	FileMenu.AddText("发送文件", keys.CmdOrCtrl("o"), func(cd *menu.CallbackData) {
		fmt.Println("发送文件")
	})
	FileMenu.AddSeparator()
	FileMenu.AddText("退出", keys.CmdOrCtrl("q"), func(cd *menu.CallbackData) {
		runtime.Quit(app.ctx)
	})

	// View menu
	ViewMenu := AppMenu.AddSubmenu("视图")
	ViewMenu.AddText("刷新", keys.Key("f5"), func(cd *menu.CallbackData) {
		fmt.Println("刷新")
	})

	// Help menu
	HelpMenu := AppMenu.AddSubmenu("帮助")
	HelpMenu.AddText("关于", nil, func(cd *menu.CallbackData) {
		runtime.MessageDialog(app.ctx, runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "关于 LanGive",
			Message: "LanGive v1.0.0\n\n一款基于 mDNS 协议的跨平台局域网文件传输工具",
		})
	})

	return AppMenu
}
