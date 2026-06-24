// 审计批量写入器(P2-15):取代"每写请求起一个 goroutine 单条 INSERT"。
// 单后台 goroutine 从缓冲 channel 消费,按数量或时间批量落库;
// 请求侧仅做一次非阻塞 channel 投递(满则丢弃计数,绝不阻塞请求、绝不无界起 goroutine)。
package middleware

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/ptis/backend/internal/db"
	"github.com/rs/zerolog/log"
)

// AuditSink 是批量写入器的落库出口(由 *db.AuditRepository 实现)。
type AuditSink interface {
	CreateBatch(ctx context.Context, items []db.AuditCreateInput) error
}

const (
	auditBufferSize   = 1024            // 缓冲容量:突发写入的吸收上限
	auditBatchSize    = 100             // 单批最多落库条数
	auditFlushEvery   = 2 * time.Second // 周期性 flush,保证低峰期也及时落库
	auditWriteTimeout = 5 * time.Second // 单批落库超时
)

type AuditWriter struct {
	ch      chan db.AuditCreateInput
	done    chan struct{}
	sink    AuditSink
	dropped atomic.Int64
}

// NewAuditWriter 启动后台写入 goroutine。调用方需在关闭时调用 Stop 排空缓冲。
func NewAuditWriter(sink AuditSink) *AuditWriter {
	w := &AuditWriter{
		ch:   make(chan db.AuditCreateInput, auditBufferSize),
		done: make(chan struct{}),
		sink: sink,
	}
	go w.run()
	return w
}

// enqueue 非阻塞投递。缓冲满则丢弃并计数——审计不应拖慢业务请求,
// 丢弃远优于无界起 goroutine 拖垮进程(本次重构正是为消除后者)。
func (w *AuditWriter) enqueue(in db.AuditCreateInput) {
	if w == nil {
		return // 未配置写入器时静默跳过(测试装配或降级场景)
	}
	select {
	case w.ch <- in:
	default:
		if n := w.dropped.Add(1); n == 1 || n%100 == 0 {
			log.Warn().Int64("dropped_total", n).Msg("审计缓冲区满,丢弃审计记录")
		}
	}
}

func (w *AuditWriter) run() {
	defer close(w.done)
	ticker := time.NewTicker(auditFlushEvery)
	defer ticker.Stop()
	buf := make([]db.AuditCreateInput, 0, auditBatchSize)

	flush := func() {
		if len(buf) == 0 {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), auditWriteTimeout)
		if err := w.sink.CreateBatch(ctx, buf); err != nil {
			log.Error().Err(err).Int("n", len(buf)).Msg("批量写审计失败")
		}
		cancel()
		buf = buf[:0]
	}

	for {
		select {
		case in, ok := <-w.ch:
			if !ok { // channel 已关闭:排空剩余 + 落库后退出
				flush()
				return
			}
			buf = append(buf, in)
			if len(buf) >= auditBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

// Stop 关闭输入、排空缓冲并落库,阻塞至完成或 ctx 超时。
// 调用前提:HTTP 服务已 Shutdown(在途请求已全部 enqueue 完毕),故不会再有写入,
// 关闭 channel 安全。
func (w *AuditWriter) Stop(ctx context.Context) {
	close(w.ch)
	select {
	case <-w.done:
	case <-ctx.Done():
		log.Warn().Msg("审计写入器关闭超时,缓冲中的记录可能未落库")
	}
}
