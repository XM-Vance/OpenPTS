// WebSocket 双向通道：/api/v1/ws/echo（JWT via query token）。
// SSE 用于服务端 → 客户端单向高频推送（告警/审批/任务），
// 这个 WS 是给客户端 → 服务端的双向交互（如实时协作、心跳上报）。
package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// wsDevMode 判断是否放行所有 Origin（开发期）。ENVIRONMENT 非 prod 即视为开发。
// 与 router 的 CORS(!IsProd()) 判定口径一致，避免 WS/CORS 规则分叉。
func wsDevMode() bool {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("ENV")
	}
	return env != "prod" && env != "production"
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 复用 CORS 同款 Origin 校验（白名单 + 同源）。
		// 开发期放行所有 Origin 便于本地调试；生产期严格校验，防 CSWSH。
		return middleware.OriginAllowed(wsDevMode(), r.Header.Get("Origin"), r.Host)
	},
}

type WebSocketHandler struct {
	jwt *auth.JWTService
}

func NewWebSocketHandler(jwt *auth.JWTService) *WebSocketHandler {
	return &WebSocketHandler{jwt: jwt}
}

type wsClientMsg struct {
	Type    string          `json:"type"`              // ping / chat / state
	Payload json.RawMessage `json:"payload,omitempty"`
}

type wsServerMsg struct {
	Type    string `json:"type"`
	From    string `json:"from,omitempty"`
	Message string `json:"message,omitempty"`
	Ts      int64  `json:"ts"`
}

// Echo GET /api/v1/ws/echo(凭证:query token 或登录 Cookie)
// 演示：收到任意消息 → 在前面加 `[echo]` 回发；每 30 秒服务端发一次心跳。
func (h *WebSocketHandler) Echo(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		token, _ = c.Cookie(auth.TokenCookieName)
	}
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少登录凭证"})
		return
	}
	claims, err := h.jwt.Parse(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token 无效"})
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("WS upgrade 失败")
		return
	}
	defer conn.Close()

	username := claims.Username
	log.Info().Str("user", username).Msg("WS 连接建立")

	// 欢迎消息
	_ = conn.WriteJSON(wsServerMsg{
		Type:    "hello",
		From:    "server",
		Message: "已连接，欢迎 " + username,
		Ts:      time.Now().Unix(),
	})

	// 心跳 goroutine（30s 一次 ping）
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := conn.WriteJSON(wsServerMsg{
					Type: "ping",
					From: "server",
					Ts:   time.Now().Unix(),
				}); err != nil {
					return
				}
			}
		}
	}()

	// 读循环
	conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})
	for {
		var msg wsClientMsg
		if err := conn.ReadJSON(&msg); err != nil {
			close(done)
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Debug().Err(err).Msg("WS 读结束")
			}
			break
		}
		// echo 回去
		_ = conn.WriteJSON(wsServerMsg{
			Type:    "echo",
			From:    "server",
			Message: "[echo] type=" + msg.Type + " " + string(msg.Payload),
			Ts:      time.Now().Unix(),
		})
	}
}
