// 统一错误响应 helper（审查报告 4.2）。
//
// 背景：handler 层有 500+ 处 `c.JSON(status, gin.H{"error": msg})` 手写错误响应，
// 状态码/文案随意，部分直接吐 err.Error() 泄露内部信息。本包提供统一封装，
// 输出与前端 client.ts（读 err.response.data.error）一致的 `{"error": msg}` 结构。
//
// 用法（逐步替换存量，新代码一律用本 helper）：
//
//	if err != nil {
//	    respondError(c, http.StatusBadRequest, "参数错误")
//	    return
//	}
//
// 存量 500+ 处手写响应格式本就一致（都是 {"error": ...}），暂不批量替换
// （易引入回归、且无功能差异）；新代码与重构时改用本 helper 统一口径。
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
)

// respondError 写出统一错误响应：{"error": msg} + 指定状态码。
// msg 不应为空；如需带原始 err 用于日志，调用方自行 log，不回传给前端。
func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

// respondFail 便捷：400 + 错误文案。
func respondFail(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
}

// respondInternalErr 便捷：500 + 通用文案（不泄露 err 细节）。
// 调用方应同时 log 原始 err。
func respondInternalErr(c *gin.Context, msg string) {
	if msg == "" {
		msg = "操作失败，请稍后重试"
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
}

// respondOrgErr 多租户写入守卫：未选具体省（总部「全部省」活跃态）时返回 400。
// 包装 db.ErrOrgRequired 的特化处理；与原 respondOrgRequired 行为一致，
// 但走统一 respondError 口径。
func respondOrgErr(c *gin.Context, err error) bool {
	if errors.Is(err, db.ErrOrgRequired) {
		respondError(c, http.StatusBadRequest, "请先选择具体省份")
		return true
	}
	return false
}

// respondBadGateway 便捷：502 + 通用「上游不可用」文案（不泄露 err 细节）。
// 用于调 algo-service / 外部服务失败时，避免把内网地址 / 上游响应体泄露给前端。
// 调用方应同时 log 原始 err。
func respondBadGateway(c *gin.Context, msg string) {
	if msg == "" {
		msg = "算法服务暂时不可用，请稍后重试"
	}
	c.JSON(http.StatusBadGateway, gin.H{"error": msg})
}
