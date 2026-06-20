package errors

import (
	"fmt"
	"net/http"
)

const (
	CodeBadRequest   = 40000
	CodeUnauthorized = 40100
	CodeForbidden    = 40300
	CodeNotFound     = 40400
	CodeConflict     = 40900
	CodeInternal     = 50000
)

type BizError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	HTTPCode int   `json:"-"`
	Err     error  `json:"-"`
}

func (e *BizError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *BizError) Unwrap() error {
	return e.Err
}

func New(code, httpCode int, message string) *BizError {
	return &BizError{
		Code:     code,
		Message:  message,
		HTTPCode: httpCode,
	}
}

func Wrap(err error, code, httpCode int, message string) *BizError {
	return &BizError{
		Code:     code,
		Message:  message,
		HTTPCode: httpCode,
		Err:      err,
	}
}

func BadRequest(message string) *BizError {
	return New(CodeBadRequest, http.StatusBadRequest, message)
}

func BadRequestf(format string, args ...interface{}) *BizError {
	return BadRequest(fmt.Sprintf(format, args...))
}

func Unauthorized(message string) *BizError {
	return New(CodeUnauthorized, http.StatusUnauthorized, message)
}

func Unauthorizedf(format string, args ...interface{}) *BizError {
	return Unauthorized(fmt.Sprintf(format, args...))
}

func Forbidden(message string) *BizError {
	return New(CodeForbidden, http.StatusForbidden, message)
}

func Forbiddenf(format string, args ...interface{}) *BizError {
	return Forbidden(fmt.Sprintf(format, args...))
}

func NotFound(message string) *BizError {
	return New(CodeNotFound, http.StatusNotFound, message)
}

func NotFoundf(format string, args ...interface{}) *BizError {
	return NotFound(fmt.Sprintf(format, args...))
}

func Conflict(message string) *BizError {
	return New(CodeConflict, http.StatusConflict, message)
}

func Conflictf(format string, args ...interface{}) *BizError {
	return Conflict(fmt.Sprintf(format, args...))
}

func Internal(message string) *BizError {
	return New(CodeInternal, http.StatusInternalServerError, message)
}

func Internalf(format string, args ...interface{}) *BizError {
	return Internal(fmt.Sprintf(format, args...))
}

func WrapInternal(err error, message string) *BizError {
	return Wrap(err, CodeInternal, http.StatusInternalServerError, message)
}

func IsBizError(err error) (*BizError, bool) {
	if err == nil {
		return nil, false
	}
	var be *BizError
	if as(err, &be) {
		return be, true
	}
	return nil, false
}

func as(err error, target interface{}) bool {
	type asInterface interface {
		As(interface{}) bool
	}
	if e, ok := err.(asInterface); ok {
		return e.As(target)
	}
	switch t := target.(type) {
	case **BizError:
		if be, ok := err.(*BizError); ok {
			*t = be
			return true
		}
	}
	return false
}
