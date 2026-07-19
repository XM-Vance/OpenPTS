// API Key 生成与校验。
// Key 格式：openpts_<base62 串>；明文只在创建时返回一次，库存 bcrypt 哈希。
// key_prefix 存明文前缀（用于展示 + 按前缀反查），避免全表 bcrypt 比对。
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	apiKeyPrefix = "openpts_"   // Key 固定前缀，便于识别
	secretBytes  = 32           // 随机秘密字节数（base64 后约 43 字符）
	prefixLen    = 16           // key_prefix 长度（含 openpts_，取明文前 16 字符）
)

// GeneratedAPIKey 创建 Key 时返回的结构：明文（只此一次）+ 前缀（入库）+ 哈希（入库）。
type GeneratedAPIKey struct {
	Plaintext string // 完整明文 Key，只在创建响应里返回一次
	Prefix    string // 明文前缀，入库用于展示/反查
	Hash      string // 完整 Key 的 bcrypt 哈希，入库
}

// GenerateAPIKey 生成一个新 API Key（明文 + 前缀 + 哈希）。
func GenerateAPIKey() (*GeneratedAPIKey, error) {
	b := make([]byte, secretBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	// url-safe base64 去掉填充，字符集 [A-Za-z0-9_-]
	secret := strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
	plaintext := apiKeyPrefix + secret

	prefix := plaintext
	if len(prefix) > prefixLen {
		prefix = prefix[:prefixLen]
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &GeneratedAPIKey{
		Plaintext: plaintext,
		Prefix:    prefix,
		Hash:      string(hash),
	}, nil
}

// VerifyAPIKey 比对明文 Key 与哈希。复用 bcrypt（与密码同算法但独立存储）。
func VerifyAPIKey(plaintext, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}

// ExtractKeyPrefix 从完整明文 Key 取前缀（用于按前缀查记录）。
// 与 GenerateAPIKey 的 prefix 逻辑保持一致。
func ExtractKeyPrefix(plaintext string) string {
	p := strings.TrimSpace(plaintext)
	if len(p) > prefixLen {
		return p[:prefixLen]
	}
	return p
}
