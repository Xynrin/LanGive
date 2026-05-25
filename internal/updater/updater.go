package updater

import (
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
	// 仓库 atom feed，避免调用 GitHub REST API
	defaultTagsFeed     = "https://github.com/Xynrin/LanGive/releases.atom"
	defaultDownloadBase = "https://github.com/Xynrin/LanGive/releases/download"
	httpTimeout         = 15 * time.Second
)

// UpdateInfo 表示一次版本检查结果
type UpdateInfo struct {
	HasUpdate     bool   `json:"has_update"`
	LatestVersion string `json:"latest_version"`
	CurrentVersion string `json:"current_version"`
	DownloadURL   string `json:"download_url"`
	ReleaseNotes  string `json:"release_notes"`
	PublishedAt   string `json:"published_at"`
}

// Service 更新服务
type Service struct {
	currentVersion string
	feedURL        string
	downloadBase   string
	httpClient     *http.Client
}

// NewService 创建更新服务
func NewService(currentVersion string) *Service {
	return &Service{
		currentVersion: currentVersion,
		feedURL:        defaultTagsFeed,
		downloadBase:   defaultDownloadBase,
		httpClient:     &http.Client{Timeout: httpTimeout},
	}
}

// atomFeed 解析 GitHub releases.atom 的最小子集
type atomFeed struct {
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	Title     string `xml:"title"`
	Updated   string `xml:"updated"`
	ID        string `xml:"id"`
	Content   string `xml:"content"`
	Link      struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
}

// CheckUpdate 检查是否有新版本
// 通过解析仓库的 releases.atom 获取最新 tag，避免调用 GitHub REST API
func (s *Service) CheckUpdate() (*UpdateInfo, error) {
	req, err := http.NewRequest("GET", s.feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/atom+xml")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
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
		return &UpdateInfo{
			HasUpdate:      false,
			CurrentVersion: s.currentVersion,
		}, nil
	}

	latest := feed.Entries[0]
	latestTag := extractTag(latest.ID)
	if latestTag == "" {
		latestTag = strings.TrimSpace(latest.Title)
	}
	latestVersion := strings.TrimPrefix(latestTag, "v")

	info := &UpdateInfo{
		LatestVersion:  latestVersion,
		CurrentVersion: s.currentVersion,
		ReleaseNotes:   strings.TrimSpace(latest.Content),
		PublishedAt:    latest.Updated,
		HasUpdate:      compareVersions(latestVersion, s.currentVersion) > 0,
	}
	if info.HasUpdate {
		info.DownloadURL = s.assetURL(latestTag)
	}
	return info, nil
}

// extractTag 从 atom entry 的 id 字段中提取 tag 名
// id 形如 tag:github.com,2008:Repository/123/v1.0.0
func extractTag(id string) string {
	idx := strings.LastIndex(id, "/")
	if idx == -1 || idx == len(id)-1 {
		return ""
	}
	return id[idx+1:]
}

// compareVersions 比较语义化版本号，返回 1/0/-1
// 不依赖第三方 semver 库；非数字段视为 0
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

// assetURL 根据当前平台拼出对应的下载地址
// 命名规则与 .github/workflows/release.yml 中产物保持一致
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
		name = ""
	}
	if name == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", s.downloadBase, tag, name)
}

// DownloadAndInstall 下载更新包并替换当前可执行文件
// 旧版本会备份为 <exe>.bak
func (s *Service) DownloadAndInstall(url string) error {
	if url == "" {
		return fmt.Errorf("empty download url")
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate executable: %w", err)
	}

	tmpDir := filepath.Dir(exePath)
	tmpFile, err := os.CreateTemp(tmpDir, "langive-update-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		return fmt.Errorf("download status %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("chmod temp: %w", err)
	}

	backup := exePath + ".bak"
	_ = os.Remove(backup)
	if err := os.Rename(exePath, backup); err != nil {
		return fmt.Errorf("backup current: %w", err)
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		// 回滚
		_ = os.Rename(backup, exePath)
		return fmt.Errorf("install new: %w", err)
	}
	return nil
}

// Restart 重新启动当前可执行文件后退出
func (s *Service) Restart() error {
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
