package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"code": code,
		"msg":  msg,
		"data": nil,
	})
}

// APIError 包装错误信息
func APIError(c *gin.Context, code int, err error) {
	Error(c, code, err.Error())
}
