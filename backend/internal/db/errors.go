package db

import "errors"

var (
	ErrUserNotFound = errors.New("用户不存在")
	ErrOrgRequired  = errors.New("请先选择具体省份")
)
