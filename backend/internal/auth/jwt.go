// JWT 签发与解析（HS256）。
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ClaimsContextKey 是 gin context 中存放 *Claims 的 key。
const ClaimsContextKey = "auth.claims"

// TokenCookieName 登录态 Cookie 名(httpOnly;/auth/login 写入、/auth/logout 清除,
// JWT 中间件与 SSE/WS handler 读取)。
const TokenCookieName = "ptis_token"

type Claims struct {
	UserID   uuid.UUID `json:"uid"`
	Username string    `json:"uname"`
	OrgID    string    `json:"org,omitempty"` // 主/默认活跃省（组织 ID）
	IsHQ     bool      `json:"hq,omitempty"`  // 总部标记：可访问/切换全部省
	jwt.RegisteredClaims
}

type JWTService struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTService(secret string, ttl time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), ttl: ttl}
}

// TTL 返回签发有效期(登录 Set-Cookie 的 Max-Age 与之对齐)。
func (s *JWTService) TTL() time.Duration { return s.ttl }

// Sign 签发 token。orgID=主活跃省，isHQ=总部标记。
func (s *JWTService) Sign(userID uuid.UUID, username, orgID string, isHQ bool) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		OrgID:    orgID,
		IsHQ:     isHQ,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ptis-backend",
			Subject:   userID.String(),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.secret)
}

// Parse 解析并校验 token。
func (s *JWTService) Parse(tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("不支持的签名方法")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("token 无效")
	}
	return claims, nil
}
