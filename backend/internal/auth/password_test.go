package auth

import "testing"

func TestHashAndCheckPasswordRoundTrip(t *testing.T) {
	plain := "Sup3r-Sec_R3t!"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword 失败: %v", err)
	}
	if hash == plain {
		t.Fatal("hash 不应等于明文")
	}
	if !CheckPassword(plain, hash) {
		t.Error("正确密码应校验通过")
	}
	if CheckPassword("WrongPassword", hash) {
		t.Error("错误密码不应校验通过")
	}
}

func TestHashPasswordProducesDistinctHashes(t *testing.T) {
	plain := "samePassword"
	h1, _ := HashPassword(plain)
	h2, _ := HashPassword(plain)
	if h1 == h2 {
		t.Error("bcrypt 应每次产生不同 salt → 哈希值应不同")
	}
	if !CheckPassword(plain, h1) || !CheckPassword(plain, h2) {
		t.Error("两个 hash 都应能验证原明文")
	}
}

func TestCheckPasswordRejectsEmpty(t *testing.T) {
	hash, _ := HashPassword("real")
	if CheckPassword("", hash) {
		t.Error("空密码不应通过")
	}
	if CheckPassword("real", "") {
		t.Error("空哈希不应通过")
	}
}
