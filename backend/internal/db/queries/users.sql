-- 阶段 1.4 鉴权会用到的查询。当前仅放示例，sqlc 生成代码后阶段 1.4 直接使用。

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = $1
  AND is_active = TRUE
LIMIT 1;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1
  AND is_active = TRUE
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, password_hash, display_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeactivateUser :exec
UPDATE users
SET is_active = FALSE
WHERE id = $1;
