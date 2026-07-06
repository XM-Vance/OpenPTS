// Origin 校验：供 CORS 与 WebSocket CheckOrigin 复用，避免规则分叉。
//
// 规则（dev 模式放行所有 origin，便于本地调试；生产模式严格校验）：
//   - 无 Origin 头：同源请求（浏览器不发 Origin）→ 放行。
//   - Origin 在 ALLOWED_ORIGINS 白名单内 → 放行。
//   - Origin 的 host 与请求 Host 头一致（同源）→ 放行。
//   - 其余 → 拒绝（CSWSH / CSRF 防护）。
package middleware

import (
	"net/url"
	"os"
	"strings"
)

// loadAllowedOrigins 解析 ALLOWED_ORIGINS（逗号分隔，去空白）为集合。
// 不缓存：每次实时读环境变量，便于测试动态修改白名单；CORS 每请求调用一次，
// 解析开销可忽略（split 一个环境变量）。
func loadAllowedOrigins() map[string]struct{} {
	set := make(map[string]struct{})
	for _, o := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			set[o] = struct{}{}
		}
	}
	return set
}

// OriginAllowed 判断请求 r 的 Origin 是否可被接受。dev 模式直接放行。
func OriginAllowed(devMode bool, origin, host string) bool {
	if devMode {
		return true
	}
	if origin == "" {
		// 同源请求浏览器不附 Origin（或附但被部分代理剥离），放行。
		return true
	}
	if _, ok := loadAllowedOrigins()[origin]; ok {
		return true
	}
	// 同源校验：Origin 的 scheme+host(+port) 与请求 Host 是否一致。
	if u, err := url.Parse(origin); err == nil {
		// u.Host 形如 "example.com:443"；对比 Host 头（含端口）。
		originHost := u.Hostname()
		originPort := u.Port()
		reqHost := host
		// 从 Host 头（可能含端口）拆出 hostname/port
		reqHostname, reqPort := splitHostPort(reqHost)
		if originHost == reqHostname {
			// 端口也要对齐：Origin 未带端口时按 scheme 默认（https=443, http=80）
			if originPort == "" || originPort == reqPort ||
				(originPort == "443" && reqPort == "" && u.Scheme == "https") ||
				(originPort == "80" && reqPort == "" && u.Scheme == "http") {
				return true
			}
		}
	}
	return false
}

// splitHostPort 把 "host:port" 拆为 hostname 与 port（无端口时 port 为空）。
func splitHostPort(h string) (string, string) {
	// 兼容 IPv6 字面量
	if strings.HasPrefix(h, "[") {
		if idx := strings.LastIndex(h, "]"); idx >= 0 {
			hostname := h[:idx+1]
			rest := h[idx+1:]
			if strings.HasPrefix(rest, ":") {
				return hostname, rest[1:]
			}
			return hostname, ""
		}
	}
	if idx := strings.LastIndex(h, ":"); idx >= 0 {
		return h[:idx], h[idx+1:]
	}
	return h, ""
}
