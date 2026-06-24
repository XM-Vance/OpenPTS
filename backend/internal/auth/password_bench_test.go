package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// 用法：go test -bench BenchmarkBcrypt -benchtime=3x ./internal/auth/...
// 输出对比不同 cost 下哈希与校验耗时，作为登录性能预算依据。

func BenchmarkBcryptHash_Cost10(b *testing.B) { benchHash(b, 10) }
func BenchmarkBcryptHash_Cost11(b *testing.B) { benchHash(b, 11) }
func BenchmarkBcryptHash_Cost12(b *testing.B) { benchHash(b, 12) }
func BenchmarkBcryptHash_Cost13(b *testing.B) { benchHash(b, 13) }

func BenchmarkBcryptCheck_Cost10(b *testing.B) { benchCheck(b, 10) }
func BenchmarkBcryptCheck_Cost11(b *testing.B) { benchCheck(b, 11) }
func BenchmarkBcryptCheck_Cost12(b *testing.B) { benchCheck(b, 12) }
func BenchmarkBcryptCheck_Cost13(b *testing.B) { benchCheck(b, 13) }

func benchHash(b *testing.B, cost int) {
	pw := []byte("Sup3r-Sec_R3t!")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := bcrypt.GenerateFromPassword(pw, cost); err != nil {
			b.Fatal(err)
		}
	}
}

func benchCheck(b *testing.B, cost int) {
	pw := []byte("Sup3r-Sec_R3t!")
	hash, err := bcrypt.GenerateFromPassword(pw, cost)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := bcrypt.CompareHashAndPassword(hash, pw); err != nil {
			b.Fatal(err)
		}
	}
}
