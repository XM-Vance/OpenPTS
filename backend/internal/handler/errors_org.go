package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ptis/backend/internal/db"
)

// respondOrgRequired 统一处理「未选择具体省份就写业务数据」的场景。
// 当 err 为 db.ErrOrgRequired（总部处于「全部省」活跃态下执行省隔离写入时由 repo 层返回）时，
// 写出 400 与中文提示并返回 true；否则返回 false，交由调用方按通用 500 处理。
//
// 用法：紧跟在 repo 写入返回 err 之后：
//
//	if err != nil {
//	    if respondOrgRequired(c, err) {
//	        return
//	    }
//	    // ……原有 500 处理
//	}
func respondOrgRequired(c *gin.Context, err error) bool {
	if errors.Is(err, db.ErrOrgRequired) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先选择具体省份"})
		return true
	}
	return false
}
