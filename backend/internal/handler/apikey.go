// 用户 API Key handler：创建/列表/吊销（登录即可）+ exchange（凭 Key 换 JWT，公开）。
// 明文 Key 只在创建时返回一次；exchange 用 Key 反查用户、签发该用户的 JWT。
package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

type APIKeyHandler struct {
	keys  *db.APIKeyRepository
	users *db.UserRepository
	jwt   *auth.JWTService
}

func NewAPIKeyHandler(keys *db.APIKeyRepository, users *db.UserRepository, jwt *auth.JWTService) *APIKeyHandler {
	return &APIKeyHandler{keys: keys, users: users, jwt: jwt}
}

// Create POST /api/v1/auth/api-keys {name} → 创建 Key，明文只返回这一次。
func (h *APIKeyHandler) Create(c *gin.Context) {
	uid := claimsUserID(c)
	if uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	_ = c.ShouldBindJSON(&req) // name 可空
	if req.Name == "" {
		req.Name = "default"
	}

	gen, err := auth.GenerateAPIKey()
	if err != nil {
		log.Error().Err(err).Msg("生成 API Key 失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成失败"})
		return
	}
	id, err := h.keys.Create(c.Request.Context(), *uid, req.Name, gen.Prefix, gen.Hash)
	if err != nil {
		log.Error().Err(err).Msg("存储 API Key 失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	log.Info().Str("user", uid.String()).Str("key_id", id.String()).Msg("API Key 已创建")
	c.JSON(http.StatusCreated, gin.H{
		"id":         id,
		"key":        gen.Plaintext, // 明文，只此一次
		"prefix":     gen.Prefix,
		"name":       req.Name,
		"created_at": time.Now(),
	})
}

// List GET /api/v1/auth/api-keys → 列出我的 Key（脱敏）
func (h *APIKeyHandler) List(c *gin.Context) {
	uid := claimsUserID(c)
	if uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	list, err := h.keys.ListByUser(c.Request.Context(), *uid)
	if err != nil {
		log.Error().Err(err).Msg("查询 API Key 列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Revoke DELETE /api/v1/auth/api-keys/:id → 吊销我的 Key
func (h *APIKeyHandler) Revoke(c *gin.Context) {
	uid := claimsUserID(c)
	if uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 key id"})
		return
	}
	if err := h.keys.Revoke(c.Request.Context(), id, *uid); err != nil {
		if errors.Is(err, db.ErrAPIKeyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Key 不存在或已吊销"})
			return
		}
		log.Error().Err(err).Msg("吊销 API Key 失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已吊销"})
}

// Exchange POST /api/v1/auth/api-key/exchange {api_key} → 凭 Key 换 JWT（公开路由）。
// Key 可放 body {api_key} 或 Authorization: Bearer <key>。签发该用户身份的 JWT。
func (h *APIKeyHandler) Exchange(c *gin.Context) {
	apiKey := c.GetString("api_key") // 由 APIKeyAuth 中间件解析填入
	if apiKey == "" {
		// 兜底：也从 body 取（中间件没挂时）
		var req struct {
			APIKey string `json:"api_key"`
		}
		_ = c.ShouldBindJSON(&req)
		apiKey = req.APIKey
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少 API Key"})
		return
	}

	prefix := auth.ExtractKeyPrefix(apiKey)
	cands, err := h.keys.GetByPrefix(c.Request.Context(), prefix)
	if err != nil {
		log.Error().Err(err).Msg("反查 API Key 失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	// 逐条 bcrypt 校验，取匹配的那条
	var matched *db.UserAPIKeyWithHash
	for _, k := range cands {
		if auth.VerifyAPIKey(apiKey, k.KeyHash) {
			matched = k
			break
		}
	}
	if matched == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API Key 无效"})
		return
	}
	// 过期检查
	if matched.ExpiresAt != nil && time.Now().After(*matched.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API Key 已过期"})
		return
	}

	// 取用户（可能已禁用）
	u, err := h.users.GetByID(c.Request.Context(), matched.UserID)
	if err != nil || u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API Key 对应用户不可用"})
		return
	}
	if !u.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户已被禁用"})
		return
	}

	orgID := ""
	if u.OrgID != nil {
		orgID = *u.OrgID
	}
	tok, err := h.jwt.Sign(u.ID, u.Username, orgID, u.IsHQ)
	if err != nil {
		log.Error().Err(err).Msg("签发 token 失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "签发 token 失败"})
		return
	}

	_ = h.keys.UpdateLastUsed(c.Request.Context(), matched.ID)
	log.Info().Str("user", u.ID.String()).Str("key_id", matched.ID.String()).Msg("API Key 换取 JWT 成功")
	c.JSON(http.StatusOK, gin.H{
		"token":    tok,
		"user_id":  u.ID,
		"username": u.Username,
	})
}

// ExtractAPIKey 是一个轻量 gin 中间件：从 Authorization: Bearer ptis_xxx 提取 API Key，
// 放进 context "api_key"。仅用于 exchange 端点（它走公开路由，不走 JWT 中间件）。
func ExtractAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			c.Set("api_key", strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		}
		c.Next()
	}
}
