package updater

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTagsFeed     = "https://github.com/Xynrin/LanGive/releases.atom"
	defaultDownloadBase = "https://github.com/Xynrin/LanGive/releases/download"
	httpTimeout         = 3 * time.Second
)

// UpdateInfo 一次版本检查结果
type UpdateInfo struct {
	HasUpdate      bool   `json:"has_update"`
	LatestVersion  string `json:"latest_version"`
	CurrentVersion string `json:"current_version"`
	DownloadURL    string `json:"download_url"`
	ReleaseNotes   string `json:"release_notes"`
	PublishedAt    string `json:"published_at"`
}

type Service struct {
	currentVersion string
	feedURL        string
	downloadBase   string
	httpClient     *http.Client
}

func NewService(currentVersion string) *Service {
	return &Service{
		currentVersion: currentVersion,
		feedURL:        defaultTagsFeed,
		downloadBase:   defaultDownloadBase,
		httpClient:     &http.Client{Timeout: httpTimeout},
	}
}

type atomFeed struct {
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title   string `xml:"title"`
	Updated string `xml:"updated"`
	ID      string `xml:"id"`
	Content string `xml:"content"`
	Link    struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
}

// CheckUpdate 检查最新版本
// 网络错误（断网/超时）静默降级为"无更新"，HTTP 非 200 仍 return error
func (s *Service) CheckUpdate() (*UpdateInfo, error) {
	req, err := http.NewRequest("GET", s.feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/atom+xml")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// 离线 / DNS 失败 / 超时 → 静默
		return &UpdateInfo{HasUpdate: false, CurrentVersion: s.currentVersion}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read feed: %w", err)
	}

	var feed atomFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}
	if len(feed.Entries) == 0 {
		return &UpdateInfo{HasUpdate: false, CurrentVersion: s.currentVersion}, nil
	}

	latest := feed.Entries[0]
	tag := extractTag(latest.ID)
	if tag == "" {
		tag = strings.TrimSpace(latest.Title)
	}
	latestVersion := strings.TrimPrefix(tag, "v")

	info := &UpdateInfo{
		LatestVersion:  latestVersion,
		CurrentVersion: s.currentVersion,
		ReleaseNotes:   strings.TrimSpace(latest.Content),
		PublishedAt:    latest.Updated,
		HasUpdate:      compareVersions(latestVersion, s.currentVersion) > 0,
	}
	if info.HasUpdate {
		info.DownloadURL = s.assetURL(tag)
	}
	return info, nil
}

func extractTag(id string) string {
	idx := strings.LastIndex(id, "/")
	if idx == -1 || idx == len(id)-1 {
		return ""
	}
	return id[idx+1:]
}

func compareVersions(a, b string) int {
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		var ai, bi int
		if i < len(pa) {
			ai, _ = strconv.Atoi(stripNonDigits(pa[i]))
		}
		if i < len(pb) {
			bi, _ = strconv.Atoi(stripNonDigits(pb[i]))
		}
		if ai > bi {
			return 1
		}
		if ai < bi {
			return -1
		}
	}
	return 0
}

func stripNonDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if b.Len() > 0 {
			break
		}
	}
	return b.String()
}

func (s *Service) assetURL(tag string) string {
	var name string
	switch runtime.GOOS {
	case "windows":
		name = fmt.Sprintf("LanGive-windows-%s.exe", runtime.GOARCH)
	case "darwin":
		name = fmt.Sprintf("LanGive-macos-%s.zip", runtime.GOARCH)
	case "linux":
		name = fmt.Sprintf("LanGive-linux-%s", runtime.GOARCH)
	default:
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", s.downloadBase, tag, name)
}

// DownloadAndInstall 下载并安装更新，按平台分流
func (s *Service) DownloadAndInstall(url string) error {
	if url == "" {
		return fmt.Errorf("empty download url")
	}
	switch runtime.GOOS {
	case "windows":
		return s.installWindows(url)
	case "darwin":
		return s.installDarwin(url)
	default:
		return s.installLinux(url)
	}
}

// downloadTo 下载 url 到 dst，返回大小
func (s *Service) downloadTo(url, dst string) error {
	// 下载阶段不受 httpTimeout 限制
	client := &http.Client{Timeout: 0}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download status %d", resp.StatusCode)
	}
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return fmt.Errorf("write: %w", err)
	}
	return f.Close()
}

// installLinux 直接 rename 替换二进制 + .bak 备份
func (s *Service) installLinux(url string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(exePath), "langive-update-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	if err := s.downloadTo(url, tmpPath); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}
	backup := exePath + ".bak"
	_ = os.Remove(backup)
	if err := os.Rename(exePath, backup); err != nil {
		return fmt.Errorf("backup: %w", err)
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		_ = os.Rename(backup, exePath)
		return fmt.Errorf("install: %w", err)
	}
	return nil
}

// installWindows 用 .bat 脚本延迟替换，绕过运行中 .exe 文件锁
func (s *Service) installWindows(url string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)

	newPath := filepath.Join(exeDir, fmt.Sprintf("langive-update-%d.exe", time.Now().UnixNano()))
	if err := s.downloadTo(url, newPath); err != nil {
		return err
	}

	batPath := filepath.Join(exeDir, "langive-update.bat")
	bat := fmt.Sprintf(`@echo off
:loop
tasklist /FI "IMAGENAME eq %s" 2>NUL | find /I "%s" >NUL
if "%%ERRORLEVEL%%"=="0" (
    timeout /t 1 /nobreak >NUL
    goto loop
)
move /Y "%s" "%s" >NUL
start "" "%s"
del "%%~f0"
`, exeName, exeName, newPath, exePath, exePath)
	if err := os.WriteFile(batPath, []byte(bat), 0644); err != nil {
		os.Remove(newPath)
		return fmt.Errorf("write bat: %w", err)
	}

	cmd := exec.Command("cmd", "/C", "start", "", batPath)
	if err := cmd.Start(); err != nil {
		os.Remove(newPath)
		os.Remove(batPath)
		return fmt.Errorf("launch updater script: %w", err)
	}
	return nil
}

// installDarwin 解压 zip 替换 .app bundle
func (s *Service) installDarwin(url string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	// exePath 通常是 <App>/LanGive.app/Contents/MacOS/LanGive
	macosDir := filepath.Dir(exePath)
	contentsDir := filepath.Dir(macosDir)
	appBundle := filepath.Dir(contentsDir)
	if filepath.Ext(appBundle) != ".app" {
		return fmt.Errorf("not running from .app bundle: %s", appBundle)
	}
	parentDir := filepath.Dir(appBundle)

	zipPath := filepath.Join(parentDir, fmt.Sprintf("langive-update-%d.zip", time.Now().UnixNano()))
	defer os.Remove(zipPath)
	if err := s.downloadTo(url, zipPath); err != nil {
		return err
	}

	extractDir, err := os.MkdirTemp(parentDir, "langive-update-")
	if err != nil {
		return fmt.Errorf("mktemp: %w", err)
	}
	defer os.RemoveAll(extractDir)
	if err := unzipDir(zipPath, extractDir); err != nil {
		return fmt.Errorf("unzip: %w", err)
	}

	newApp := filepath.Join(extractDir, "LanGive.app")
	if st, err := os.Stat(newApp); err != nil || !st.IsDir() {
		return fmt.Errorf("zip missing LanGive.app at top level")
	}

	backup := appBundle + ".bak"
	_ = os.RemoveAll(backup)
	if err := os.Rename(appBundle, backup); err != nil {
		return fmt.Errorf("backup app: %w", err)
	}
	if err := os.Rename(newApp, appBundle); err != nil {
		_ = os.Rename(backup, appBundle)
		return fmt.Errorf("install app: %w", err)
	}
	return nil
}

// unzipDir 解压 zipPath 到 destDir，防 zip slip
func unzipDir(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		target := filepath.Join(destDir, f.Name)
		absTarget, err := filepath.Abs(target)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(absTarget, absDest+string(os.PathSeparator)) && absTarget != absDest {
			return fmt.Errorf("zip slip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return err
			}
			continue
		}
		if f.Mode()&os.ModeSymlink != 0 {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			link, _ := io.ReadAll(rc)
			rc.Close()
			os.MkdirAll(filepath.Dir(target), 0755)
			if err := os.Symlink(string(link), target); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			rc.Close()
			out.Close()
			return err
		}
		rc.Close()
		out.Close()
	}
	return nil
}

// Restart 重启程序
// Windows 上不 fork 自身（.bat 会接管），其他平台 fork+exit
func (s *Service) Restart() error {
	if runtime.GOOS == "windows" {
		go func() {
			time.Sleep(200 * time.Millisecond)
			os.Exit(0)
		}()
		return nil
	}
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}
	cmd := exec.Command(exePath, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("restart: %w", err)
	}
	go func() {
		time.Sleep(200 * time.Millisecond)
		os.Exit(0)
	}()
	return nil
}
