package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	status  int
	code    string
	message string
	cause   error
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *AppError) Unwrap() error {
	return e.cause
}

func (e *AppError) HTTPStatus() int {
	return e.status
}

func (e *AppError) Code() string {
	return e.code
}

func (e *AppError) Message() string {
	return e.message
}

func (e *AppError) Cause() error {
	return e.cause
}

func New(cause error, status int, code string, message string) *AppError {
	return &AppError{status: status, code: code, message: message, cause: cause}
}

func Newf(status int, code string, format string, args ...interface{}) *AppError {
	return &AppError{status: status, code: code, message: fmt.Sprintf(format, args...)}
}

func Wrap(cause error, status int, code string, message string) *AppError {
	return &AppError{status: status, code: code, message: message, cause: cause}
}

func NotFound(resource string, cause error) *AppError {
	return &AppError{status: http.StatusNotFound, code: "NOT_FOUND", message: resource + " not found", cause: cause}
}

func NotFoundf(resource string, cause error, format string, args ...interface{}) *AppError {
	msg := fmt.Sprintf(format, args...)
	return &AppError{status: http.StatusNotFound, code: "NOT_FOUND", message: msg, cause: cause}
}

func Unauthorized(msg string, cause error) *AppError {
	return &AppError{status: http.StatusUnauthorized, code: "UNAUTHORIZED", message: msg, cause: cause}
}

func Forbidden(msg string, cause error) *AppError {
	return &AppError{status: http.StatusForbidden, code: "FORBIDDEN", message: msg, cause: cause}
}

func BadRequest(msg string, cause error) *AppError {
	return &AppError{status: http.StatusBadRequest, code: "BAD_REQUEST", message: msg, cause: cause}
}

func Validation(msg string, cause error) *AppError {
	return &AppError{status: http.StatusBadRequest, code: "VALIDATION_FAILED", message: msg, cause: cause}
}

func Conflict(msg string, cause error) *AppError {
	return &AppError{status: http.StatusConflict, code: "CONFLICT", message: msg, cause: cause}
}

func Internal(msg string, cause error) *AppError {
	return &AppError{status: http.StatusInternalServerError, code: "INTERNAL_ERROR", message: msg, cause: cause}
}

func RateLimited(msg string) *AppError {
	return &AppError{status: http.StatusTooManyRequests, code: "RATE_LIMITED", message: msg}
}

func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus()
	}
	return http.StatusInternalServerError
}

func GetCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code()
	}
	return "INTERNAL_ERROR"
}

func GetMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message()
	}
	return err.Error()
}

func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

func IsNotFound(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.HTTPStatus() == http.StatusNotFound
}

func IsUnauthorized(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.HTTPStatus() == http.StatusUnauthorized
}

func IsForbidden(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.HTTPStatus() == http.StatusForbidden
}

func IsValidation(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.HTTPStatus() == http.StatusBadRequest
}
