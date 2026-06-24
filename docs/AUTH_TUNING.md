# 鉴权性能调优

## bcrypt cost 选型

bcrypt cost 越高安全性越强，但单次哈希/校验耗时指数增长。
登录端点延迟主要由此决定，需要在 **安全性** 与 **用户体验** 之间平衡。

### 实测数据（Apple M2，2026-05）

| Cost | 哈希耗时 | 校验耗时 |
|---|---|---|
| 10 | ~60 ms | ~58 ms |
| 11 | ~123 ms | ~123 ms |
| 12 | ~239 ms | ~235 ms |
| 13 | ~479 ms | ~472 ms |

每升 1 级耗时翻倍，符合预期。

### 建议

| 环境 | cost | 说明 |
|---|---|---|
| dev / CI | 10 | 默认值，单元测试快 |
| 生产（< 1k 用户） | 12 | OWASP 推荐下限；登录 ~250 ms 用户感知良好 |
| 高敏感生产 | 13 | 银行/金融场景；登录 ~500 ms |
| 千万级用户 | 11 + 速率限制 + 行为风控 | 用业务层补强单点 cost |

### 调整方法

`internal/auth/password.go` 中的 `HashPassword`：

```go
func HashPassword(plain string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
    // bcrypt.DefaultCost = 10
    return string(hash), err
}
```

升级到 cost 12：

```go
const passwordCost = 12

func HashPassword(plain string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(plain), passwordCost)
    return string(hash), err
}
```

bcrypt 哈希自带 cost 元信息，**升级 cost 不影响存量用户登录**（旧密码仍能校验通过）。
若想强制存量用户重哈希，可在 Login handler 检测到 cost < 当前 cost 时透明重写：

```go
if !auth.CheckPassword(req.Password, u.PasswordHash) { ... }
// 透明升级
if cost, _ := bcrypt.Cost([]byte(u.PasswordHash)); cost < passwordCost {
    newHash, _ := auth.HashPassword(req.Password)
    _ = users.UpdatePassword(ctx, u.ID, newHash)
}
```

## 限流配合

cost 越高，登录端点越脆弱（CPU 密集）。建议同时启用：

- `middleware.StrictRateLimit(0.2, 3)` 防爆破（已配置）
- nginx 层 `limit_req_zone` 二级保护
- 失败 N 次锁定账户（待加）

## JWT TTL

- access token：24h（当前）
- refresh token：14d（待加）
- 长会话场景：滑动续期（每次请求 < 30 min 内续期）

## 验证

```bash
cd backend && go test -bench BenchmarkBcrypt -benchtime=3x -run '^$' ./internal/auth/...
```
