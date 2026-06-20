package response

import (
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
	ErrorWithCode(c, http.StatusBadRequest, CodeError, message)
}

func Unauthorized(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusUnauthorized, CodeError, message)
}

func Forbidden(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusForbidden, CodeError, message)
}

func NotFound(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusNotFound, CodeError, message)
}

func InternalError(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusInternalServerError, CodeError, message)
}
