// 种子工具：创建默认管理员账号，并自动分配 super_admin 角色（前提：已执行 0015_auth_seed 迁移）。
// super_admin 还会被置为总部（is_hq=true）——可访问/切换全部省，也能写业务数据。
// （迁移 0051 的「现有 super_admin → is_hq」回填在 migrate 时跑，命不中 migrate 之后才创建的 seed 用户。）
// 用法：
//   make seed
//   或：go run ./cmd/seed [-username admin] [-display 管理员] [-role super_admin]
//   指定密码：go run ./cmd/seed -password MySecret123
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/ptis/backend/internal/auth"
	"github.com/ptis/backend/internal/config"
	"github.com/ptis/backend/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// generatePassword 生成 16 字节随机十六进制密码（32 字符）。
func generatePassword() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成随机密码失败: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func main() {
	username := flag.String("username", "admin", "用户名")
	password := flag.String("password", "", "密码（留空则自动生成随机密码）")
	displayName := flag.String("display", "管理员", "显示名称")
	roleCode := flag.String("role", "super_admin", "自动分配的角色编码（角色不存在则跳过）")
	flag.Parse()

	if *password == "" {
		p, err := generatePassword()
		if err != nil {
			log.Fatalf("%v", err)
		}
		*password = p
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("连接 Postgres 失败: %v", err)
	}
	defer pool.Close()

	repo := db.NewUserRepository(pool)
	hash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("哈希密码失败: %v", err)
	}

	var userID uuid.UUID

	u, err := repo.Create(ctx, *username, hash, *displayName)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			fmt.Printf("用户 '%s' 已存在，跳过创建\n", *username)
			existing, errGet := repo.GetByUsername(ctx, *username)
			if errGet != nil {
				log.Fatalf("读取已存在用户失败: %v", errGet)
			}
			userID = existing.ID
		} else {
			log.Fatalf("创建用户失败: %v", err)
		}
	} else {
		fmt.Printf("已创建用户: username=%s id=%s password=%s\n", u.Username, u.ID, *password)
		userID = u.ID
	}

	// 自动分配角色（前提：0015_auth_seed 已 migrate up）
	if *roleCode != "" {
		ok, err := repo.AssignRoleIfRoleExists(ctx, userID, *roleCode)
		switch {
		case err != nil:
			fmt.Printf("分配角色失败: %v\n", err)
		case ok:
			fmt.Printf("已分配角色: %s\n", *roleCode)
		default:
			fmt.Printf("角色 '%s' 不存在或未激活，跳过（请确认 migrate 已执行到 0015_auth_seed）\n", *roleCode)
		}
	}

	// 默认让 super_admin 成为总部（HQ），可切换/访问全部省。
	// 原因：迁移 0051「现有 super_admin → is_hq」只在 migrate 时回填当时已存在的用户，
	// 而 seed 在 migrate 之后才创建 admin，0051 的 UPDATE 命中不到——故 seed 侧补一刀，
	// 保证开箱即有一个能看全部省、也能写业务数据的管理员（否则 HQ「全部省」下写业务会 400）。
	if *roleCode == "super_admin" {
		if err := repo.SetUserOrgs(ctx, userID, nil, true, ""); err != nil {
			fmt.Printf("置为总部失败: %v\n", err)
		} else {
			fmt.Printf("已置为总部（is_hq=true）\n")
		}
	}

	fmt.Println("登录方式：POST /api/v1/auth/login")
}
