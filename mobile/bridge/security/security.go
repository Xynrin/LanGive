// Package security 是 internal/security 的 gomobile 友好包装。
package security

import (
	isec "github.com/Xynrin/LanGive/internal/security"
)

// Manager 包装 internal/security.Manager。
type Manager struct {
	inner *isec.Manager
}

// NewSecurityManager 与 internal API 同名。
func NewSecurityManager() *Manager {
	return &Manager{inner: isec.NewSecurityManager()}
}

// CreatePublicSession 创建公共会话。
func (m *Manager) CreatePublicSession() {
	m.inner.CreatePublicSession()
}

// CreatePrivateSession 创建隐私会话。
func (m *Manager) CreatePrivateSession() {
	m.inner.CreatePrivateSession()
}

// JoinSession 加入会话。
func (m *Manager) JoinSession(sessionID, deviceUUID string) error {
	return m.inner.JoinSession(sessionID, deviceUUID)
}

// LeaveSession 离开会话。
func (m *Manager) LeaveSession(sessionID, deviceUUID string) {
	m.inner.LeaveSession(sessionID, deviceUUID)
}
