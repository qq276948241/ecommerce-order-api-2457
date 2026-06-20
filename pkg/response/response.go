package response

import (
	"ecommerce-backend/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess = 0
	CodeError   = 1
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func Fail(c *gin.Context, err error) {
	if be, ok := errors.IsBizError(err); ok {
		c.JSON(be.HTTPCode, Response{
			Code:    be.Code,
			Message: be.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, Response{
		Code:    errors.CodeInternal,
		Message: "服务器内部错误",
	})
}

func Error(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeError,
		Message: message,
	})
}

func ErrorWithCode(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	Fail(c, errors.BadRequest(message))
}

func Unauthorized(c *gin.Context, message string) {
	Fail(c, errors.Unauthorized(message))
}

func Forbidden(c *gin.Context, message string) {
	Fail(c, errors.Forbidden(message))
}

func NotFound(c *gin.Context, message string) {
	Fail(c, errors.NotFound(message))
}

func InternalError(c *gin.Context, message string) {
	Fail(c, errors.Internal(message))
}
