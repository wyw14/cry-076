package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wyw14/cry-076/internal/domain"
	appmw "github.com/wyw14/cry-076/internal/middleware"
)

type APIError struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Fields    []FieldError `json:"fields,omitempty"`
	RequestID string       `json:"request_id"`
}
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondError(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "资源不存在"
	case errors.Is(err, domain.ErrForbidden):
		status, code, message = http.StatusForbidden, "FORBIDDEN", "没有操作权限"
	case errors.Is(err, domain.ErrConflict):
		status, code, message = http.StatusConflict, "VERSION_CONFLICT", "资源版本已变化"
	case errors.Is(err, domain.ErrIdempotencyReuse):
		status, code, message = http.StatusConflict, "IDEMPOTENCY_REUSED", "幂等键对应了不同请求"
	case errors.Is(err, domain.ErrInvalidTransition):
		status, code, message = http.StatusUnprocessableEntity, "INVALID_TRANSITION", "当前状态不允许此操作"
	case errors.Is(err, domain.ErrValidation):
		status, code, message = http.StatusBadRequest, "VALIDATION_FAILED", "请求参数不符合规则"
	}
	c.JSON(status, gin.H{"error": APIError{Code: code, Message: message, RequestID: appmw.GetRequestID(c)}})
}
func respondBindingError(c *gin.Context, err error) {
	fields := make([]FieldError, 0)
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, item := range validationErrors {
			fields = append(fields, FieldError{Field: item.Field(), Code: item.Tag(), Message: "字段校验失败"})
		}
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": APIError{Code: "VALIDATION_FAILED", Message: "请求参数不符合规则", Fields: fields, RequestID: appmw.GetRequestID(c)}})
}
func respondData(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data, "request_id": appmw.GetRequestID(c)})
}
func respondList(c *gin.Context, data any, total, page, pageSize int) {
	c.JSON(http.StatusOK, gin.H{"data": data, "page": gin.H{"number": page, "size": pageSize, "total": total}, "request_id": appmw.GetRequestID(c)})
}
