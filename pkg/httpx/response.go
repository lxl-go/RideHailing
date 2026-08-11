package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ride-hailing/pkg/errors"
)

// Response 统一 HTTP 响应结构（对标文档 4.3 节错误处理）
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	TraceID string      `json:"trace_id,omitempty"`
}

// Success 成功响应（对标文档空值处理规范）
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Data:    data,
		Message: "success",
	})
}

// Paginated 分页响应（对标文档列表接口规范：items 为空数组而非 null）
func Paginated(c *gin.Context, list interface{}, total int64) {
	if list == nil {
		list = []interface{}{}
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    gin.H{"list": list, "total": total},
		"message": "success",
	})
}

// Error 错误响应（对标文档错误返回格式）
func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.Error); ok {
		c.JSON(appErr.HTTPStatus(), Response{
			Code:    appErr.HTTPStatus(),
			Data:    nil,
			Message: appErr.Message,
			TraceID: c.GetString("trace_id"),
		})
		return
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    500,
		Data:    nil,
		Message: "系统繁忙",
		TraceID: c.GetString("trace_id"),
	})
}

// RequestID 从 Gin Context 获取请求 ID（对标文档 ctx 透传）
func RequestID(c *gin.Context) string {
	if id, ok := c.Get("request_id"); ok {
		return id.(string)
	}
	return ""
}
