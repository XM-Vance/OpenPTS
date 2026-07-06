package middleware

import (
	"testing"
)

// 验证失败计数 → 锁定 → 解锁 流转。
func TestLoginLockoutCountsAndLocks(t *testing.T) {
	l := newLoginLockoutForTest()
	key := "1.2.3.4:alice"

	// 前 4 次失败不锁定
	for i := 0; i < loginLockoutThreshold-1; i++ {
		l.RecordFailure(key)
		if l.IsLocked(key) {
			t.Fatalf("失败 %d 次不应锁定", i+1)
		}
	}

	// 第 5 次失败：锁定
	l.RecordFailure(key)
	if !l.IsLocked(key) {
		t.Fatal("失败达阈值应锁定")
	}
}

// 成功登录清零计数，避免历史失败累积误锁。
func TestLoginLockoutSuccessResets(t *testing.T) {
	l := newLoginLockoutForTest()
	key := "1.2.3.4:bob"

	l.RecordFailure(key)
	l.RecordFailure(key)
	l.RecordFailure(key) // 3 次
	l.RecordSuccess(key) // 成功清零

	// 再失败 3 次不应锁定（计数已重置，阈值 5）
	l.RecordFailure(key)
	l.RecordFailure(key)
	l.RecordFailure(key)
	if l.IsLocked(key) {
		t.Fatal("成功后计数清零，3 次失败不应锁定")
	}
}

// 不同 key 互不影响（IP+username 组合隔离）。
func TestLoginLockoutKeyIsolation(t *testing.T) {
	l := newLoginLockoutForTest()
	for i := 0; i < loginLockoutThreshold; i++ {
		l.RecordFailure("10.0.0.1:carol")
	}
	if !l.IsLocked("10.0.0.1:carol") {
		t.Fatal("carol 应被锁定")
	}
	if l.IsLocked("10.0.0.1:dave") {
		t.Fatal("dave 不应因 carol 被锁")
	}
	if l.IsLocked("10.0.0.2:carol") {
		t.Fatal("不同 IP 的 carol 不应被锁")
	}
}
