// 写入回归冒烟测试（repo 层，带真实库）。
//
// 守护 PR #25 迁移 0068 那类回归：documents.uploaded_by / approval_requests.submitted_by /
// reviewed_by / document_applies.applied_by 被改为 UUID FK，写入路径若传用户名字符串则运行时
// 报「invalid input syntax for type uuid」500。这里直接以真实用户 UUID 调 repo 写入，断言不报错
// 且落库为 UUID——比 HTTP 层冒烟更稳（不依赖 worker / registry / SSE / 鉴权）。
//
// CI：backend job 在「migrate up」之后运行 go test，故能命中真库。
// 本地：未设置 DATABASE_URL 或库不可用时自动 Skip。
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ptis/backend/internal/db"
)

// smokePool 连接 CI 提供的 Postgres；未配置/连不上则 Skip。
func smokePool(t *testing.T) *db.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL 未设置，跳过写入冒烟测试")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pool, err := db.New(ctx, dsn)
	if err != nil {
		t.Skipf("数据库不可用（%v），跳过写入冒烟测试", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("数据库不可 Ping（%v），跳过写入冒烟测试", err)
	}
	return pool
}

// smokeFirstOrgID 取一个省份（组织）id，供 OrgFilter 上下文。
func smokeFirstOrgID(t *testing.T, pool *db.Pool) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(),
		`SELECT id::text FROM organizations LIMIT 1`).Scan(&id); err != nil {
		t.Skipf("无法获取 org_id（库未迁移/无组织）: %v", err)
	}
	return id
}

// smokeUser 建一个测试用户，满足 *_by 的外键（migrate-only 的库无 seed 用户）。
func smokeUser(t *testing.T, pool *db.Pool, orgID string) (uuid.UUID, func()) {
	t.Helper()
	id := uuid.New()
	uname := fmt.Sprintf("smoke_%d", time.Now().UnixNano())
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO users (id, username, password_hash, org_id, is_hq, is_active)
		 VALUES ($1, $2, 'x', $3::uuid, true, true)`, id, uname, orgID); err != nil {
		t.Skipf("无法创建测试用户: %v", err)
	}
	return id, func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	}
}

// TestSmoke_DocumentUploadedByUUID 直测 documents.uploaded_by 的 UUID 写入。
func TestSmoke_DocumentUploadedByUUID(t *testing.T) {
	pool := smokePool(t)
	defer pool.Close()
	orgID := smokeFirstOrgID(t, pool)
	userID, cleanupUser := smokeUser(t, pool, orgID)
	defer cleanupUser()

	ctx := db.WithOrg(context.Background(), orgID)
	repo := db.NewDocumentRepository(pool)
	doc, err := repo.Create(ctx, "smoke.txt", "text/plain", 5, uuid.NewString(), "csv", "", &userID)
	if err != nil {
		t.Fatalf("文档 Create 失败（uploaded_by UUID 写入回归？）: %v", err)
	}
	defer pool.Exec(context.Background(), `DELETE FROM documents WHERE id = $1`, doc.ID)

	if doc.UploadedBy == nil || *doc.UploadedBy != userID.String() {
		t.Errorf("uploaded_by 应为 %s，得到 %v", userID, doc.UploadedBy)
	}
}

// TestSmoke_ApprovalSubmittedReviewedByUUID 直测 submitted_by / reviewed_by 的 UUID 写入。
func TestSmoke_ApprovalSubmittedReviewedByUUID(t *testing.T) {
	pool := smokePool(t)
	defer pool.Close()
	orgID := smokeFirstOrgID(t, pool)
	userID, cleanupUser := smokeUser(t, pool, orgID)
	defer cleanupUser()

	ctx := db.WithOrg(context.Background(), orgID)
	repo := db.NewApprovalRepository(pool)

	a, err := repo.Create(ctx, db.ApprovalInput{
		Resource:    "smoke",
		ResourceID:  "smoke-" + uuid.NewString(),
		Title:       "冒烟测试审批",
		Payload:     json.RawMessage(`{}`),
		SubmittedBy: userID.String(),
	})
	if err != nil {
		t.Fatalf("审批 Create 失败（submitted_by UUID 写入回归？）: %v", err)
	}
	defer pool.Exec(context.Background(), `DELETE FROM approval_requests WHERE id = $1`, a.ID)

	if a.SubmittedBy != userID.String() {
		t.Errorf("submitted_by 应为 %s，得到 %q", userID, a.SubmittedBy)
	}

	a2, err := repo.Transition(ctx, a.ID, "approved", userID.String(), "冒烟通过")
	if err != nil {
		t.Fatalf("审批 Transition 失败（reviewed_by UUID 写入回归？）: %v", err)
	}
	if a2.ReviewedBy == nil || *a2.ReviewedBy != userID.String() {
		t.Errorf("reviewed_by 应为 %s，得到 %v", userID, a2.ReviewedBy)
	}
}
