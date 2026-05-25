package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// TokenGenerator 令牌生成器
type TokenGenerator struct {
	secretKey []byte
}

// NewTokenGenerator 创建令牌生成器
func NewTokenGenerator(secret string) *TokenGenerator {
	hash := sha256.Sum256([]byte(secret))
	return &TokenGenerator{
		secretKey: hash[:],
	}
}

// GenerateTransferToken 生成传输令牌
// 用于验证发送方和接收方之间的连接
func (tg *TokenGenerator) GenerateTransferToken(senderUUID, receiverUUID string) string {
	data := fmt.Sprintf("%s:%s", senderUUID, receiverUUID)
	h := hmac.New(sha256.New, tg.secretKey)
	h.Write([]byte(data))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// VerifyTransferToken 验证传输令牌
func (tg *TokenGenerator) VerifyTransferToken(senderUUID, receiverUUID, token string) bool {
	expected := tg.GenerateTransferToken(senderUUID, receiverUUID)
	return hmac.Equal([]byte(expected), []byte(token))
}

// GenerateFileChecksum 生成文件校验和
func (tg *TokenGenerator) GenerateFileChecksum(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
