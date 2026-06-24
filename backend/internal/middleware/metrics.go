// Prometheus 指标中间件：HTTP 请求数 + 耗时直方图，按 method/route/status 维度。
// 端点 /metrics 由 PrometheusHandler() 暴露，无需鉴权。
package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ptis_http_requests_total",
			Help: "HTTP 请求总数",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ptis_http_request_duration_seconds",
			Help:    "HTTP 请求耗时直方图（秒）",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "route", "status"},
	)

	httpInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ptis_http_in_flight_requests",
		Help: "当前进行中的 HTTP 请求数",
	})
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, httpInFlight)
}

// Metrics gin 中间件：每个请求上下两次打点。
// route 用 gin 注册的模式（FullPath）而非真实路径，避免高基数（如 :id 实际值）。
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		httpInFlight.Inc()
		start := time.Now()
		c.Next()
		httpInFlight.Dec()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		labels := prometheus.Labels{
			"method": c.Request.Method,
			"route":  route,
			"status": status,
		}
		httpRequestsTotal.With(labels).Inc()
		httpRequestDuration.With(labels).Observe(time.Since(start).Seconds())
	}
}

// PrometheusHandler 返回 /metrics 处理器（gin 包装）。
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
