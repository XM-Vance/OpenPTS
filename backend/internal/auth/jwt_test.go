package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTSignAndParseRoundTrip(t *testing.T) {
	svc := NewJWTService("test-secret-at-least-32-characters-xxxx", 1*time.Hour)

	uid := uuid.New()
	token, err := svc.Sign(uid, "alice", "FJ", true)
	if err != nil {
		t.Fatalf("Sign 失败: %v", err)
	}
	if token == "" {
		t.Fatal("token 为空")
	}

	claims, err := svc.Parse(token)
	if err != nil {
		t.Fatalf("Parse 失败: %v", err)
	}
	if claims.UserID != uid {
		t.Errorf("UserID 不一致: 期望 %s 实际 %s", uid, claims.UserID)
	}
	if claims.Username != "alice" {
		t.Errorf("Username 不一致: 期望 alice 实际 %s", claims.Username)
	}
	if claims.OrgID != "FJ" || !claims.IsHQ {
		t.Errorf("org/hq 未透传: org=%q hq=%v", claims.OrgID, claims.IsHQ)
	}
}

func TestJWTRejectsWrongSecret(t *testing.T) {
	signer := NewJWTService("secret-A-at-least-32-characters-xxxxxxx", time.Hour)
	verifier := NewJWTService("secret-B-at-least-32-characters-xxxxxxx", time.Hour)

	token, err := signer.Sign(uuid.New(), "bob", "", false)
	if err != nil {
		t.Fatalf("Sign 失败: %v", err)
	}
	if _, err := verifier.Parse(token); err == nil {
		t.Error("用不同 secret 解析应失败，却成功了")
	}
}

func TestJWTRejectsExpiredToken(t *testing.T) {
	svc := NewJWTService("test-secret-at-least-32-characters-xxxx", 1*time.Millisecond)
	token, err := svc.Sign(uuid.New(), "carol", "", false)
	if err != nil {
		t.Fatalf("Sign 失败: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if _, err := svc.Parse(token); err == nil {
		t.Error("过期 token 应解析失败")
	}
}

func TestJWTRejectsTampered(t *testing.T) {
	svc := NewJWTService("test-secret-at-least-32-characters-xxxx", time.Hour)
	token, err := svc.Sign(uuid.New(), "dave", "", false)
	if err != nil {
		t.Fatalf("Sign 失败: %v", err)
	}
	// 篡改 payload 段首字符：改动签名输入（header.payload），HMAC 校验必然失败。
	// 不可只改 token 末位——签名段 base64url 末字符仅高位有效、低位是填充，改成固定
	// 'X' 有约 1/16 概率解码为同一字节，篡改失效、测试偶发假失败（曾导致 CI flaky）。
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("非预期的 JWT 结构：%d 段", len(parts))
	}
	if parts[1][0] == 'A' {
		parts[1] = "B" + parts[1][1:]
	} else {
		parts[1] = "A" + parts[1][1:]
	}
	tampered := strings.Join(parts, ".")
	if _, err := svc.Parse(tampered); err == nil {
		t.Error("被篡改的 token 应解析失败")
	}
}
