// OpenTelemetry 分布式追踪初始化。
//
// 启动时调 InitTracer 注册全局 TracerProvider（OTLP HTTP 导出），
// 后续 otelgin（HTTP 入口）自动生成/传播 trace；otelhttp 出站追踪由二次开发者按需启用。
// 连不上 collector（如本地未起监控栈）时静默失败，不影响业务。
//
// 导出目标走 env OTEL_EXPORTER_OTLP_ENDPOINT（默认 http://tempo:4318），
// 与 docker-compose.monitoring.yml 的 Tempo 服务对接。
package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace/noop"
)

// InitTracer 初始化全局 TracerProvider（OTLP HTTP 导出），返回 shutdown 函数。
//
// endpoint 走 env OTEL_EXPORTER_OTLP_ENDPOINT（默认 http://tempo:4318，
// Tempo 在监控栈启动时存在）。service.name 标为 openpts-backend。
// 连不上时降级为 noop provider，业务零影响。
//
// 用法：
//
//	shutdown, err := observability.InitTracer(ctx)
//	if err != nil { log.Warn().Err(err).Msg("OTel 初始化失败，trace 禁用") }
//	defer shutdown(ctx)
func InitTracer(ctx context.Context, serviceVersion string) (shutdown func(context.Context) error, err error) {
	// OTLP HTTP exporter，走标准 OTel env（OTEL_EXPORTER_OTLP_ENDPOINT 等）。
	// 默认指向 http://tempo:4318（docker 网络内 Tempo 服务）。
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithInsecure(), // Tempo 在内网，无 TLS
	)
	if err != nil {
		// 降级：连不上 exporter 也返回 noop，让业务继续（OTel 默认异步、失败静默）。
		otel.SetTracerProvider(noop.NewTracerProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{}))
		return func(context.Context) error { return nil }, nil
	}

	// resource 用单一 schema 构建，避免 resource.Default()（1.41.0）与
	// semconv/v1.26.0 的 SchemaURL 冲突。仅含 service.name/version 等核心属性。
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("openpts-backend"),
		semconv.ServiceVersion(serviceVersion),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter), // 异步批量导出，不阻塞业务
		sdktrace.WithResource(res),
		// 采样：AlwaysSample（追踪量不大，全采便于调试；量大后改 ParentBased(TraceIDRatioBased)）
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	// W3C traceparent 传播（跨服务透传的标准格式）
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	return tp.Shutdown, nil
}
