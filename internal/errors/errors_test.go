package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_SetsAllFields_ReturnsAppError(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := New(cause, http.StatusNotFound, "NOT_FOUND", "resource missing")

	assert.Equal(t, http.StatusNotFound, err.HTTPStatus())
	assert.Equal(t, "NOT_FOUND", err.Code())
	assert.Equal(t, "resource missing", err.Message())
	assert.Equal(t, cause, err.Cause())
	assert.Equal(t, cause, err.Unwrap())
}

func TestNew_NilCause_ReturnsAppError(t *testing.T) {
	err := New(nil, http.StatusBadRequest, "BAD_REQUEST", "invalid input")

	assert.Nil(t, err.Cause())
	assert.Nil(t, err.Unwrap())
	assert.Equal(t, "invalid input", err.Message())
}

func TestNewf_FormattedMessage_ReturnsAppError(t *testing.T) {
	err := Newf(http.StatusConflict, "CONFLICT", "item %s already exists with id %d", "widget", 42)

	assert.Equal(t, http.StatusConflict, err.HTTPStatus())
	assert.Equal(t, "CONFLICT", err.Code())
	assert.Equal(t, "item widget already exists with id 42", err.Message())
	assert.Nil(t, err.Cause())
}

func TestNewf_NoArgs_ReturnsAppError(t *testing.T) {
	err := Newf(http.StatusInternalServerError, "INTERNAL", "plain message")

	assert.Equal(t, "plain message", err.Message())
}

func TestWrap_WithCause_ReturnsAppError(t *testing.T) {
	cause := fmt.Errorf("db connection failed")
	err := Wrap(cause, http.StatusInternalServerError, "DB_ERROR", "database unavailable")

	assert.Equal(t, cause, err.Cause())
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus())
	assert.Equal(t, "DB_ERROR", err.Code())
	assert.Equal(t, "database unavailable", err.Message())
}

func TestError_WithCause_IncludesCauseInMessage(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := New(cause, http.StatusInternalServerError, "INTERNAL", "service failed")

	assert.Equal(t, "service failed: connection refused", err.Error())
}

func TestError_WithoutCause_ReturnsMessageOnly(t *testing.T) {
	err := New(nil, http.StatusBadRequest, "BAD_REQUEST", "bad input")

	assert.Equal(t, "bad input", err.Error())
}

func TestUnwrap_ReturnsUnderlyingCause(t *testing.T) {
	cause := fmt.Errorf("inner")
	err := New(cause, http.StatusOK, "OK", "msg")

	unwrapped := err.Unwrap()
	require.NotNil(t, unwrapped)
	assert.Equal(t, "inner", unwrapped.Error())
}

func TestUnwrap_NilCause_ReturnsNil(t *testing.T) {
	err := New(nil, http.StatusOK, "OK", "msg")
	assert.Nil(t, err.Unwrap())
}

func TestNotFound_SetsStatusAndCode(t *testing.T) {
	cause := fmt.Errorf("query returned 0 rows")
	err := NotFound("user", cause)

	assert.Equal(t, http.StatusNotFound, err.HTTPStatus())
	assert.Equal(t, "NOT_FOUND", err.Code())
	assert.Equal(t, "user not found", err.Message())
	assert.Equal(t, cause, err.Cause())
}

func TestNotFound_NilCause_ReturnsAppError(t *testing.T) {
	err := NotFound("session", nil)

	assert.Equal(t, http.StatusNotFound, err.HTTPStatus())
	assert.Equal(t, "session not found", err.Message())
	assert.Nil(t, err.Cause())
}

func TestNotFoundf_FormattedMessage_ReturnsAppError(t *testing.T) {
	cause := fmt.Errorf("gone")
	err := NotFoundf("doc", cause, "document %s version %d missing", "readme", 3)

	assert.Equal(t, http.StatusNotFound, err.HTTPStatus())
	assert.Equal(t, "NOT_FOUND", err.Code())
	assert.Equal(t, "document readme version 3 missing", err.Message())
	assert.Equal(t, cause, err.Cause())
}

func TestUnauthorized_SetsStatusAndCode(t *testing.T) {
	cause := fmt.Errorf("expired token")
	err := Unauthorized("invalid credentials", cause)

	assert.Equal(t, http.StatusUnauthorized, err.HTTPStatus())
	assert.Equal(t, "UNAUTHORIZED", err.Code())
	assert.Equal(t, "invalid credentials", err.Message())
	assert.Equal(t, cause, err.Cause())
}

func TestForbidden_SetsStatusAndCode(t *testing.T) {
	err := Forbidden("access denied", nil)

	assert.Equal(t, http.StatusForbidden, err.HTTPStatus())
	assert.Equal(t, "FORBIDDEN", err.Code())
	assert.Equal(t, "access denied", err.Message())
}

func TestBadRequest_SetsStatusAndCode(t *testing.T) {
	err := BadRequest("missing field", nil)

	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus())
	assert.Equal(t, "BAD_REQUEST", err.Code())
	assert.Equal(t, "missing field", err.Message())
}

func TestValidation_SetsStatusAndCode(t *testing.T) {
	err := Validation("email is invalid", nil)

	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus())
	assert.Equal(t, "VALIDATION_FAILED", err.Code())
	assert.Equal(t, "email is invalid", err.Message())
}

func TestConflict_SetsStatusAndCode(t *testing.T) {
	err := Conflict("duplicate entry", nil)

	assert.Equal(t, http.StatusConflict, err.HTTPStatus())
	assert.Equal(t, "CONFLICT", err.Code())
	assert.Equal(t, "duplicate entry", err.Message())
}

func TestInternal_SetsStatusAndCode(t *testing.T) {
	cause := fmt.Errorf("nil pointer")
	err := Internal("something broke", cause)

	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus())
	assert.Equal(t, "INTERNAL_ERROR", err.Code())
	assert.Equal(t, "something broke", err.Message())
	assert.Equal(t, cause, err.Cause())
}

func TestRateLimited_SetsStatusAndCode(t *testing.T) {
	err := RateLimited("too many requests")

	assert.Equal(t, http.StatusTooManyRequests, err.HTTPStatus())
	assert.Equal(t, "RATE_LIMITED", err.Code())
	assert.Equal(t, "too many requests", err.Message())
	assert.Nil(t, err.Cause())
}

func TestGetHTTPStatus_AppError_ReturnsStatus(t *testing.T) {
	err := BadRequest("bad", nil)
	assert.Equal(t, http.StatusBadRequest, GetHTTPStatus(err))
}

func TestGetHTTPStatus_StandardError_Returns500(t *testing.T) {
	err := fmt.Errorf("plain error")
	assert.Equal(t, http.StatusInternalServerError, GetHTTPStatus(err))
}

func TestGetHTTPStatus_NilError_Returns500(t *testing.T) {
	assert.Equal(t, http.StatusInternalServerError, GetHTTPStatus(nil))
}

func TestGetCode_AppError_ReturnsCode(t *testing.T) {
	err := NotFound("thing", nil)
	assert.Equal(t, "NOT_FOUND", GetCode(err))
}

func TestGetCode_StandardError_ReturnsInternalError(t *testing.T) {
	err := fmt.Errorf("plain")
	assert.Equal(t, "INTERNAL_ERROR", GetCode(err))
}

func TestGetCode_NilError_PanicsOrReturnsDefault(t *testing.T) {
	assert.Equal(t, "INTERNAL_ERROR", GetCode(nil))
}

func TestGetMessage_AppError_ReturnsMessage(t *testing.T) {
	err := Conflict("clash", nil)
	assert.Equal(t, "clash", GetMessage(err))
}

func TestGetMessage_StandardError_ReturnsErrorString(t *testing.T) {
	err := fmt.Errorf("std error msg")
	assert.Equal(t, "std error msg", GetMessage(err))
}

func TestIsAppError_AppError_ReturnsTrue(t *testing.T) {
	err := BadRequest("x", nil)
	assert.True(t, IsAppError(err))
}

func TestIsAppError_WrappedAppError_ReturnsTrue(t *testing.T) {
	appErr := BadRequest("x", nil)
	wrapped := fmt.Errorf("outer: %w", appErr)
	assert.True(t, IsAppError(wrapped))
}

func TestIsAppError_StandardError_ReturnsFalse(t *testing.T) {
	assert.False(t, IsAppError(fmt.Errorf("plain")))
}

func TestIsAppError_NilError_ReturnsFalse(t *testing.T) {
	assert.False(t, IsAppError(nil))
}

func TestIsNotFound_CorrectStatus_ReturnsTrue(t *testing.T) {
	err := NotFound("user", nil)
	assert.True(t, IsNotFound(err))
}

func TestIsNotFound_WrongStatus_ReturnsFalse(t *testing.T) {
	err := BadRequest("bad", nil)
	assert.False(t, IsNotFound(err))
}

func TestIsNotFound_StandardError_ReturnsFalse(t *testing.T) {
	assert.False(t, IsNotFound(fmt.Errorf("not an AppError")))
}

func TestIsNotFound_NilError_ReturnsFalse(t *testing.T) {
	assert.False(t, IsNotFound(nil))
}

func TestIsUnauthorized_CorrectStatus_ReturnsTrue(t *testing.T) {
	err := Unauthorized("nope", nil)
	assert.True(t, IsUnauthorized(err))
}

func TestIsUnauthorized_WrongStatus_ReturnsFalse(t *testing.T) {
	err := Forbidden("nope", nil)
	assert.False(t, IsUnauthorized(err))
}

func TestIsUnauthorized_StandardError_ReturnsFalse(t *testing.T) {
	assert.False(t, IsUnauthorized(fmt.Errorf("no")))
}

func TestIsUnauthorized_NilError_ReturnsFalse(t *testing.T) {
	assert.False(t, IsUnauthorized(nil))
}

func TestIsForbidden_CorrectStatus_ReturnsTrue(t *testing.T) {
	err := Forbidden("denied", nil)
	assert.True(t, IsForbidden(err))
}

func TestIsForbidden_WrongStatus_ReturnsFalse(t *testing.T) {
	err := Unauthorized("denied", nil)
	assert.False(t, IsForbidden(err))
}

func TestIsForbidden_StandardError_ReturnsFalse(t *testing.T) {
	assert.False(t, IsForbidden(fmt.Errorf("x")))
}

func TestIsForbidden_NilError_ReturnsFalse(t *testing.T) {
	assert.False(t, IsForbidden(nil))
}

func TestIsValidation_CorrectStatus_ReturnsTrue(t *testing.T) {
	err := Validation("bad email", nil)
	assert.True(t, IsValidation(err))
}

func TestIsValidation_BadRequestSameStatus_ReturnsTrue(t *testing.T) {
	err := BadRequest("bad input", nil)
	assert.True(t, IsValidation(err))
}

func TestIsValidation_WrongStatus_ReturnsFalse(t *testing.T) {
	err := NotFound("x", nil)
	assert.False(t, IsValidation(err))
}

func TestIsValidation_StandardError_ReturnsFalse(t *testing.T) {
	assert.False(t, IsValidation(fmt.Errorf("x")))
}

func TestIsValidation_NilError_ReturnsFalse(t *testing.T) {
	assert.False(t, IsValidation(nil))
}

func TestErrorsAs_UnwrapsAppError(t *testing.T) {
	inner := New(nil, http.StatusNotFound, "NOT_FOUND", "gone")
	wrapped := fmt.Errorf("layer2: %w", inner)

	var appErr *AppError
	require.True(t, errors.As(wrapped, &appErr))
	assert.Equal(t, http.StatusNotFound, appErr.HTTPStatus())
	assert.Equal(t, "gone", appErr.Message())
}

func TestErrorsAs_DoubleWrapped_UnwrapsToAppError(t *testing.T) {
	inner := BadRequest("bad", nil)
	layer1 := fmt.Errorf("l1: %w", inner)
	layer2 := fmt.Errorf("l2: %w", layer1)

	var appErr *AppError
	require.True(t, errors.As(layer2, &appErr))
	assert.Equal(t, "bad", appErr.Message())
}

func TestErrorsAs_StandardError_ReturnsFalse(t *testing.T) {
	err := fmt.Errorf("plain")
	var appErr *AppError
	assert.False(t, errors.As(err, &appErr))
}

func TestGetHTTPStatus_WrappedAppError_ReturnsStatus(t *testing.T) {
	appErr := Unauthorized("no", nil)
	wrapped := fmt.Errorf("handler: %w", appErr)
	assert.Equal(t, http.StatusUnauthorized, GetHTTPStatus(wrapped))
}

func TestGetCode_WrappedAppError_ReturnsCode(t *testing.T) {
	appErr := Conflict("dup", nil)
	wrapped := fmt.Errorf("service: %w", appErr)
	assert.Equal(t, "CONFLICT", GetCode(wrapped))
}

func TestGetMessage_WrappedAppError_ReturnsMessage(t *testing.T) {
	appErr := Forbidden("no access", nil)
	wrapped := fmt.Errorf("middleware: %w", appErr)
	assert.Equal(t, "no access", GetMessage(wrapped))
}

func TestPredicate_WrappedAppError_MatchesStatus(t *testing.T) {
	appErr := NotFound("resource", nil)
	wrapped := fmt.Errorf("handler: %w", appErr)
	assert.True(t, IsNotFound(wrapped))
}

func TestAppError_AllConvenienceConstructors_HaveDistinctCodes(t *testing.T) {
	cases := []struct {
		name     string
		err      *AppError
		status   int
		code     string
		message  string
		hasCause bool
	}{
		{"NotFound", NotFound("x", nil), http.StatusNotFound, "NOT_FOUND", "x not found", false},
		{"Unauthorized", Unauthorized("x", nil), http.StatusUnauthorized, "UNAUTHORIZED", "x", false},
		{"Forbidden", Forbidden("x", nil), http.StatusForbidden, "FORBIDDEN", "x", false},
		{"BadRequest", BadRequest("x", nil), http.StatusBadRequest, "BAD_REQUEST", "x", false},
		{"Validation", Validation("x", nil), http.StatusBadRequest, "VALIDATION_FAILED", "x", false},
		{"Conflict", Conflict("x", nil), http.StatusConflict, "CONFLICT", "x", false},
		{"Internal", Internal("x", nil), http.StatusInternalServerError, "INTERNAL_ERROR", "x", false},
		{"RateLimited", RateLimited("x"), http.StatusTooManyRequests, "RATE_LIMITED", "x", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.status, tc.err.HTTPStatus())
			assert.Equal(t, tc.code, tc.err.Code())
			assert.Equal(t, tc.message, tc.err.Message())
			if tc.hasCause {
				assert.NotNil(t, tc.err.Cause())
			} else {
				assert.Nil(t, tc.err.Cause())
			}
		})
	}
}

func TestNew_WithCause_ErrorIncludesCause(t *testing.T) {
	cause := fmt.Errorf("timeout")
	err := New(cause, http.StatusGatewayTimeout, "TIMEOUT", "request timed out")

	assert.Equal(t, "request timed out: timeout", err.Error())
}

func TestWrap_IsIdenticalToNew(t *testing.T) {
	cause := fmt.Errorf("root")
	w := Wrap(cause, 400, "B", "msg")
	n := New(cause, 400, "B", "msg")

	assert.Equal(t, w.HTTPStatus(), n.HTTPStatus())
	assert.Equal(t, w.Code(), n.Code())
	assert.Equal(t, w.Message(), n.Message())
	assert.Equal(t, w.Error(), n.Error())
}
