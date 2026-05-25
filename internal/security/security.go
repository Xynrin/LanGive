package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const (
	// TokenValidityDuration 令牌有效期
	TokenValidityDuration = 24 * time.Hour
)

type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	IsPublic  bool      `json:"is_public"`
	Devices   []string  `json:"devices"` // 设备UUID列表
}

type DeviceToken struct {
	DeviceUUID   string    `json:"device_uuid"`
	Token        string    `json:"token"`
	TokenHash    string    `json:"token_hash"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastVerified time.Time `json:"last_verified"`
}

type Manager struct {
	sessions map[string]*Session      // 会话ID -> Session
	tokens   map[string]*DeviceToken  // TokenHash -> DeviceToken
	devices  map[string]*DeviceToken  // DeviceUUID -> DeviceToken

	sessionsMux sync.RWMutex
	tokensMux   sync.RWMutex
}

func NewSecurityManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		tokens:   make(map[string]*DeviceToken),
		devices:  make(map[string]*DeviceToken),
	}
}

// CreatePublicSession 创建公共会话
func (m *Manager) CreatePublicSession() *Session {
	m.sessionsMux.Lock()
	defer m.sessionsMux.Unlock()

	session := &Session{
		ID:        "public",
		Name:      "公共会话",
		CreatedAt: time.Now(),
		IsPublic:  true,
		Devices:   make([]string, 0),
	}
	m.sessions["public"] = session
	return session
}

// CreatePrivateSession 创建私有会话
func (m *Manager) CreatePrivateSession() *Session {
	m.sessionsMux.Lock()
	defer m.sessionsMux.Unlock()

	session := &Session{
		ID:        generateSessionID(),
		Name:      "私有会话",
		CreatedAt: time.Now(),
		IsPublic:  false,
		Devices:   make([]string, 0),
	}
	m.sessions[session.ID] = session
	return session
}

// GetSession 获取会话
func (m *Manager) GetSession(sessionID string) *Session {
	m.sessionsMux.RLock()
	defer m.sessionsMux.RUnlock()
	return m.sessions[sessionID]
}

// JoinSession 设备加入会话
func (m *Manager) JoinSession(sessionID, deviceUUID string) error {
	m.sessionsMux.Lock()
	defer m.sessionsMux.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// 检查设备是否已在会话中
	for _, d := range session.Devices {
		if d == deviceUUID {
			return nil
		}
	}
	session.Devices = append(session.Devices, deviceUUID)
	return nil
}

// LeaveSession 设备离开会话
func (m *Manager) LeaveSession(sessionID, deviceUUID string) {
	m.sessionsMux.Lock()
	defer m.sessionsMux.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return
	}

	for i, d := range session.Devices {
		if d == deviceUUID {
			session.Devices = append(session.Devices[:i], session.Devices[i+1:]...)
			break
		}
	}
}

// GenerateToken 为设备生成连接令牌
func (m *Manager) GenerateToken(deviceUUID string) (*DeviceToken, error) {
	m.tokensMux.Lock()
	defer m.tokensMux.Unlock()

	// 生成随机令牌
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// 计算令牌的哈希值用于存储
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	now := time.Now()
	dt := &DeviceToken{
		DeviceUUID:   deviceUUID,
		Token:        token,
		TokenHash:    tokenHash,
		IssuedAt:     now,
		ExpiresAt:    now.Add(TokenValidityDuration),
		LastVerified: now,
	}

	m.tokens[tokenHash] = dt
	m.devices[deviceUUID] = dt

	return dt, nil
}

// VerifyToken 验证设备令牌
func (m *Manager) VerifyToken(deviceUUID, token string) bool {
	m.tokensMux.Lock()
	defer m.tokensMux.Unlock()

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	dt, ok := m.tokens[tokenHash]
	if !ok {
		return false
	}

	// 检查设备UUID是否匹配
	if dt.DeviceUUID != deviceUUID {
		return false
	}

	// 检查是否过期
	if time.Now().After(dt.ExpiresAt) {
		return false
	}

	// 更新最后验证时间
	dt.LastVerified = time.Now()
	return true
}

// RevokeToken 撤销设备令牌
func (m *Manager) RevokeToken(deviceUUID string) {
	m.tokensMux.Lock()
	defer m.tokensMux.Unlock()

	dt, ok := m.devices[deviceUUID]
	if !ok {
		return
	}

	delete(m.tokens, dt.TokenHash)
	delete(m.devices, deviceUUID)
}

// GetSessionDevices 获取会话中的所有设备
func (m *Manager) GetSessionDevices(sessionID string) []string {
	m.sessionsMux.RLock()
	defer m.sessionsMux.RUnlock()

	session := m.sessions[sessionID]
	if session == nil {
		return nil
	}

	devices := make([]string, len(session.Devices))
	copy(devices, session.Devices)
	return devices
}

// CleanupExpiredTokens 清理过期的令牌
func (m *Manager) CleanupExpiredTokens() {
	m.tokensMux.Lock()
	defer m.tokensMux.Unlock()

	now := time.Now()
	for hash, dt := range m.tokens {
		if now.After(dt.ExpiresAt) {
			delete(m.tokens, hash)
			delete(m.devices, dt.DeviceUUID)
		}
	}
}

// StartCleanupRoutine 启动定时清理过期令牌的goroutine
func (m *Manager) StartCleanupRoutine(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			m.CleanupExpiredTokens()
		}
	}()
}

func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
