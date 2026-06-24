// Postgres 连接池封装（基于 pgx/v5）。
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool 主写库连接池；嵌入 pgxpool.Pool 让所有 repo 直接调用 .Query/.Exec/.QueryRow。
// 当配置了 replica DSN 时，replica 字段非空，read-only 查询可以走 ReadPool()。
type Pool struct {
	*pgxpool.Pool
	replica *pgxpool.Pool
}

func openPool(ctx context.Context, dsn string, maxConns int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("解析 DSN 失败: %w", err)
	}
	cfg.MaxConns = maxConns
	cfg.MinConns = 5 // R7: 保持 5 个热连接，减少冷启动延迟
	cfg.MaxConnLifetime = 5 * time.Minute // R7: 5 分钟回收，防止长连接累积问题
	cfg.MaxConnIdleTime = 15 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("创建连接池失败: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping Postgres 失败: %w", err)
	}
	return pool, nil
}

// New 创建主连接池并执行连通性测试。
func New(ctx context.Context, dsn string) (*Pool, error) {
	master, err := openPool(ctx, dsn, 25) // R7: MaxOpen=25
	if err != nil {
		return nil, err
	}
	return &Pool{Pool: master}, nil
}

// WithReplica 在主连接池基础上挂载只读副本连接池；replicaDSN 为空时直接返回。
// 副本不可用时不阻塞启动，仅打 warn 日志（由调用方处理）。
func (p *Pool) WithReplica(ctx context.Context, replicaDSN string) error {
	if replicaDSN == "" {
		return nil
	}
	rep, err := openPool(ctx, replicaDSN, 15) // R7: 副本 MaxConns=15
	if err != nil {
		return err
	}
	p.replica = rep
	return nil
}

// ReadPool 返回只读副本（若有），否则回退到主库。
// 用法：repo 在「明确只读」的查询里用 r.pool.ReadPool().Query(...)；写操作仍用 r.pool（master）。
func (p *Pool) ReadPool() interface {
	Query(ctx context.Context, sql string, args ...any) (pgxRows, error)
} {
	if p.replica != nil {
		return readAdapter{p.replica}
	}
	return readAdapter{p.Pool}
}

// HealthCheck 供健康检查接口调用。
func (p *Pool) HealthCheck(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return p.Ping(pingCtx)
}

// Close 关闭主库与副本。
func (p *Pool) Close() {
	if p.replica != nil {
		p.replica.Close()
	}
	p.Pool.Close()
}

// ReplicaInUse 标记是否启用了副本（供 /health 自检显示）。
func (p *Pool) ReplicaInUse() bool { return p.replica != nil }

// 适配 ReadPool 返回类型，避免对外暴露 pgxpool 细节。
type pgxRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

type readAdapter struct{ inner *pgxpool.Pool }

func (a readAdapter) Query(ctx context.Context, sql string, args ...any) (pgxRows, error) {
	return a.inner.Query(ctx, sql, args...)
}
