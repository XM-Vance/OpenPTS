package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Executor 抽象 *Pool 与 pgx.Tx 共同实现的查询能力，让需要事务原子性的
// 业务方法（如审批 Approve 的 Transition+Apply）能在同一事务内执行。
// *Pool（嵌入 *pgxpool.Pool）与 pgx.Tx 都满足此接口。
type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
