// SSE 推送：把后端事件（新告警、调度状态变化等）以 text/event-stream 推到前端。
// 路由：GET /api/v1/stream/alerts?token=<JWT>（token 用 query 是因为 EventSource 不支持自定义 header）
//
// 实现要点：
//   - 每个客户端一个 channel；hub 负责广播；客户端断开时摘除。
//   - 心跳 15s，保证防火墙 / 代理不切流。
//   - 任意 handler 可调用 SSEHub.Publish 推送事件。
package handler

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ptis/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type SSEEvent struct {
	Type string `json:"type"` // alert / job_done / info
	Data any    `json:"data"`
}

type sseClient struct {
	id  string
	ch  chan SSEEvent
}

type SSEHub struct {
	mu      sync.RWMutex
	clients map[string]*sseClient
}

func NewSSEHub() *SSEHub { return &SSEHub{clients: map[string]*sseClient{}} }

func (h *SSEHub) addClient(id string) *sseClient {
	c := &sseClient{id: id, ch: make(chan SSEEvent, 16)}
	h.mu.Lock()
	h.clients[id] = c
	h.mu.Unlock()
	return c
}

func (h *SSEHub) removeClient(id string) {
	h.mu.Lock()
	if c, ok := h.clients[id]; ok {
		close(c.ch)
		delete(h.clients, id)
	}
	h.mu.Unlock()
}

// Publish 向所有连接的客户端广播事件。channel 写阻塞时直接丢弃（避免慢消费拖死）。
func (h *SSEHub) Publish(ev SSEEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		select {
		case c.ch <- ev:
		default:
		}
	}
}

// ClientCount 当前连接数（仅供 metrics / 调试）。
func (h *SSEHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// OnlineUsers 当前在线用户列表（按 client_id 前缀 username 去重）。
func (h *SSEHub) OnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for id := range h.clients {
		// id 形如 "username-1234567890"，username 在最后一个 '-' 之前
		idx := -1
		for k := len(id) - 1; k >= 0; k-- {
			if id[k] == '-' {
				idx = k
				break
			}
		}
		uname := id
		if idx > 0 {
			uname = id[:idx]
		}
		if _, ok := seen[uname]; ok {
			continue
		}
		seen[uname] = struct{}{}
		out = append(out, uname)
	}
	return out
}

// Online GET /api/v1/online —— 返回在线用户名单与连接数。
func (h *SSEHandler) Online(c *gin.Context) {
	users := h.hub.OnlineUsers()
	c.JSON(http.StatusOK, gin.H{
		"connections": h.hub.ClientCount(),
		"users":       users,
		"count":       len(users),
	})
}

// PublishJobEvent 实现 scheduler.EventPublisher 接口。
func (h *SSEHub) PublishJobEvent(jobName, status string, durationMs int, errMsg string) {
	icon := "✅"
	if status == "failed" {
		icon = "❌"
	}
	msg := icon + " 任务「" + jobName + "」" + status
	if errMsg != "" {
		msg += "：" + errMsg
	}
	h.Publish(SSEEvent{
		Type: "job",
		Data: map[string]any{
			"job":         jobName,
			"status":      status,
			"duration_ms": durationMs,
			"error":       errMsg,
			"message":     msg,
			"ts":          time.Now().Unix(),
		},
	})
}

// SSEHandler 装配 gin Handler。
type SSEHandler struct {
	hub *SSEHub
	jwt *auth.JWTService
}

func NewSSEHandler(hub *SSEHub, jwt *auth.JWTService) *SSEHandler {
	return &SSEHandler{hub: hub, jwt: jwt}
}

// PublishTest POST /api/v1/stream/test —— 手工广播一条测试事件，便于联调。
func (h *SSEHandler) PublishTest(c *gin.Context) {
	var body struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Type == "" {
		body.Type = "alert"
	}
	if body.Message == "" {
		body.Message = "测试推送 - 你好"
	}
	h.hub.Publish(SSEEvent{
		Type: body.Type,
		Data: gin.H{
			"message": body.Message,
			"ts":      time.Now().Unix(),
		},
	})
	c.JSON(http.StatusOK, gin.H{
		"clients_notified": h.hub.ClientCount(),
		"message":          "已推送",
	})
}

// Stream GET /api/v1/stream/alerts
// EventSource 无法附 Authorization 头:凭证优先取 query token(过渡期旧客户端),
// 缺失时回退登录 Cookie(httpOnly,同源自动携带,P1-8 默认通道)。
func (h *SSEHandler) Stream(c *gin.Context) {
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

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // nginx 关闭缓冲

	clientID := fmt.Sprintf("%s-%d", claims.Username, time.Now().UnixNano())
	client := h.hub.addClient(clientID)
	defer h.hub.removeClient(clientID)

	// 首次推送一条 hello 让前端确认连接成功
	c.SSEvent("hello", gin.H{"client_id": clientID, "user": claims.Username})
	c.Writer.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case <-ticker.C:
			c.SSEvent("ping", gin.H{"t": time.Now().Unix()})
			return true
		case ev, ok := <-client.ch:
			if !ok {
				return false
			}
			c.SSEvent(ev.Type, ev.Data)
			return true
		}
	})
}
