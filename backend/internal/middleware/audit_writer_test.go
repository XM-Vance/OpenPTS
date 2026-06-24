package middleware

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ptis/backend/internal/db"
)

// fakeSink 记录每次批量落库的条数,断言无丢失 + 确实成批。
type fakeSink struct {
	mu      sync.Mutex
	total   int
	batches int
}

func (s *fakeSink) CreateBatch(_ context.Context, items []db.AuditCreateInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.total += len(items)
	s.batches++
	return nil
}

func (s *fakeSink) snapshot() (total, batches int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.total, s.batches
}

// 缓冲容量内投递的记录:Stop 后必须一条不丢全部落库,且至少成一批。
func TestAuditWriterFlushesAllOnStop(t *testing.T) {
	sink := &fakeSink{}
	w := NewAuditWriter(sink)

	const n = 250 // > auditBatchSize(100),触发按量 flush;< 缓冲容量(1024),不丢
	for i := 0; i < n; i++ {
		w.enqueue(db.AuditCreateInput{Method: "POST", Path: "/x"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	w.Stop(ctx)

	total, batches := sink.snapshot()
	if total != n {
		t.Fatalf("审计记录有丢失:落库 %d,期望 %d", total, n)
	}
	if batches < 2 {
		t.Errorf("期望成多批落库(250/100),实际批数 %d", batches)
	}
}

// nil 写入器投递应安全 no-op(测试装配/降级场景)。
func TestAuditWriterEnqueueNilSafe(t *testing.T) {
	var w *AuditWriter
	w.enqueue(db.AuditCreateInput{Method: "PUT"}) // 不 panic 即通过
}
