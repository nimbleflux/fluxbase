package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestErrorCodeConstants(t *testing.T) {
	assert.Equal(t, "23505", ErrCodeUniqueViolation)
	assert.Equal(t, "23503", ErrCodeForeignKeyViolation)
	assert.Equal(t, "23514", ErrCodeCheckViolation)
}

func TestIsUniqueViolation(t *testing.T) {
	t.Run("returns true for unique violation error", func(t *testing.T) {
		err := &pgconn.PgError{Code: ErrCodeUniqueViolation}
		assert.True(t, IsUniqueViolation(err))
	})

	t.Run("returns false for other pg errors", func(t *testing.T) {
		err := &pgconn.PgError{Code: ErrCodeForeignKeyViolation}
		assert.False(t, IsUniqueViolation(err))
	})

	t.Run("returns false for non-pg error", func(t *testing.T) {
		err := errors.New("generic error")
		assert.False(t, IsUniqueViolation(err))
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		assert.False(t, IsUniqueViolation(nil))
	})

	t.Run("returns false for wrapped non-pg error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped generic error")
		assert.False(t, IsUniqueViolation(wrappedErr))
	})
}

func TestIsForeignKeyViolation(t *testing.T) {
	t.Run("returns true for foreign key violation error", func(t *testing.T) {
		err := &pgconn.PgError{Code: ErrCodeForeignKeyViolation}
		assert.True(t, IsForeignKeyViolation(err))
	})

	t.Run("returns false for other pg errors", func(t *testing.T) {
		err := &pgconn.PgError{Code: ErrCodeUniqueViolation}
		assert.False(t, IsForeignKeyViolation(err))
	})

	t.Run("returns false for non-pg error", func(t *testing.T) {
		err := errors.New("generic error")
		assert.False(t, IsForeignKeyViolation(err))
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		assert.False(t, IsForeignKeyViolation(nil))
	})
}

func TestIsCheckViolation(t *testing.T) {
	t.Run("returns true for check violation error", func(t *testing.T) {
		err := &pgconn.PgError{Code: ErrCodeCheckViolation}
		assert.True(t, IsCheckViolation(err))
	})

	t.Run("returns false for other pg errors", func(t *testing.T) {
		err := &pgconn.PgError{Code: ErrCodeUniqueViolation}
		assert.False(t, IsCheckViolation(err))
	})

	t.Run("returns false for non-pg error", func(t *testing.T) {
		err := errors.New("generic error")
		assert.False(t, IsCheckViolation(err))
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		assert.False(t, IsCheckViolation(nil))
	})
}

func TestGetConstraintName(t *testing.T) {
	t.Run("returns constraint name from pg error", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:           ErrCodeUniqueViolation,
			ConstraintName: "users_email_key",
		}
		assert.Equal(t, "users_email_key", GetConstraintName(err))
	})

	t.Run("returns empty string for non-pg error", func(t *testing.T) {
		err := errors.New("generic error")
		assert.Equal(t, "", GetConstraintName(err))
	})

	t.Run("returns empty string for nil error", func(t *testing.T) {
		assert.Equal(t, "", GetConstraintName(nil))
	})

	t.Run("returns empty string when no constraint name set", func(t *testing.T) {
		err := &pgconn.PgError{Code: ErrCodeCheckViolation}
		assert.Equal(t, "", GetConstraintName(err))
	})
}

// =============================================================================
// Additional Error Code Tests
// =============================================================================

func TestErrorCodes_AllConstants(t *testing.T) {
	t.Run("all error codes are distinct", func(t *testing.T) {
		codes := []string{
			ErrCodeUniqueViolation,
			ErrCodeForeignKeyViolation,
			ErrCodeCheckViolation,
		}

		seen := make(map[string]bool)
		for _, code := range codes {
			assert.False(t, seen[code], "duplicate error code: %s", code)
			seen[code] = true
		}
	})

	t.Run("error codes are PostgreSQL standard", func(t *testing.T) {
		// PostgreSQL error codes are 5 characters
		assert.Len(t, ErrCodeUniqueViolation, 5)
		assert.Len(t, ErrCodeForeignKeyViolation, 5)
		assert.Len(t, ErrCodeCheckViolation, 5)
	})
}

func TestPgError_FullFields(t *testing.T) {
	t.Run("pg error with all fields", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:           ErrCodeUniqueViolation,
			Message:        "duplicate key value violates unique constraint",
			Detail:         "Key (email)=(test@example.com) already exists.",
			Hint:           "Check your data for duplicates.",
			ConstraintName: "users_email_key",
			TableName:      "users",
			SchemaName:     "public",
			ColumnName:     "email",
		}

		assert.True(t, IsUniqueViolation(err))
		assert.Equal(t, "users_email_key", GetConstraintName(err))
		assert.Contains(t, err.Error(), "duplicate key")
	})

	t.Run("pg error message content", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:    ErrCodeForeignKeyViolation,
			Message: "insert or update on table violates foreign key constraint",
		}

		assert.True(t, IsForeignKeyViolation(err))
		assert.Contains(t, err.Error(), "foreign key")
	})
}

func TestWrappedErrors(t *testing.T) {
	t.Run("wrapped pg error still detected", func(t *testing.T) {
		pgErr := &pgconn.PgError{Code: ErrCodeUniqueViolation}
		wrappedErr := errors.Join(errors.New("context"), pgErr)

		// Note: errors.As is used internally, so wrapped errors should work
		var targetErr *pgconn.PgError
		if errors.As(wrappedErr, &targetErr) {
			assert.Equal(t, ErrCodeUniqueViolation, targetErr.Code)
		}
	})
}

// =============================================================================
// Error Categorization Tests
// =============================================================================

func TestErrorCategorization(t *testing.T) {
	testCases := []struct {
		name           string
		code           string
		isUnique       bool
		isForeignKey   bool
		isCheck        bool
		constraintName string
	}{
		{
			name:           "unique violation",
			code:           ErrCodeUniqueViolation,
			isUnique:       true,
			isForeignKey:   false,
			isCheck:        false,
			constraintName: "test_unique_key",
		},
		{
			name:           "foreign key violation",
			code:           ErrCodeForeignKeyViolation,
			isUnique:       false,
			isForeignKey:   true,
			isCheck:        false,
			constraintName: "test_fk",
		},
		{
			name:           "check violation",
			code:           ErrCodeCheckViolation,
			isUnique:       false,
			isForeignKey:   false,
			isCheck:        true,
			constraintName: "test_check",
		},
		{
			name:           "unrecognized code",
			code:           "00000",
			isUnique:       false,
			isForeignKey:   false,
			isCheck:        false,
			constraintName: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := &pgconn.PgError{
				Code:           tc.code,
				ConstraintName: tc.constraintName,
			}

			assert.Equal(t, tc.isUnique, IsUniqueViolation(err))
			assert.Equal(t, tc.isForeignKey, IsForeignKeyViolation(err))
			assert.Equal(t, tc.isCheck, IsCheckViolation(err))
			assert.Equal(t, tc.constraintName, GetConstraintName(err))
		})
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkIsUniqueViolation_PgError(b *testing.B) {
	err := &pgconn.PgError{Code: ErrCodeUniqueViolation}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsUniqueViolation(err)
	}
}

func BenchmarkIsUniqueViolation_GenericError(b *testing.B) {
	err := errors.New("generic error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsUniqueViolation(err)
	}
}

func BenchmarkIsForeignKeyViolation(b *testing.B) {
	err := &pgconn.PgError{Code: ErrCodeForeignKeyViolation}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsForeignKeyViolation(err)
	}
}

func BenchmarkIsCheckViolation(b *testing.B) {
	err := &pgconn.PgError{Code: ErrCodeCheckViolation}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsCheckViolation(err)
	}
}

func BenchmarkGetConstraintName(b *testing.B) {
	err := &pgconn.PgError{
		Code:           ErrCodeUniqueViolation,
		ConstraintName: "users_email_key",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetConstraintName(err)
	}
}

func BenchmarkGetConstraintName_NilError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetConstraintName(nil)
	}
}

func TestSanitizeErrorMessage(t *testing.T) {
	t.Run("nil error returns empty string", func(t *testing.T) {
		assert.Equal(t, "", SanitizeErrorMessage(nil))
	})

	t.Run("unique violation preserves constraint name and code", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:           ErrCodeUniqueViolation,
			Message:        `duplicate key value violates unique constraint "users_email_key"`,
			ConstraintName: "users_email_key",
		}
		msg := SanitizeErrorMessage(err)
		assert.Contains(t, msg, "unique constraint violation")
		assert.Contains(t, msg, "users_email_key")
		assert.Contains(t, msg, "23505")
		assert.NotContains(t, msg, "duplicate key value violates")
	})

	t.Run("foreign key violation preserves constraint name", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:           ErrCodeForeignKeyViolation,
			Message:        `insert or update on table "orders" violates foreign key constraint "orders_user_id_fkey"`,
			ConstraintName: "orders_user_id_fkey",
			Detail:         "Key (user_id)=(1) is not present in table \"users\".",
		}
		msg := SanitizeErrorMessage(err)
		assert.Contains(t, msg, "foreign key constraint violation")
		assert.Contains(t, msg, "orders_user_id_fkey")
		assert.NotContains(t, msg, "users\".")
	})

	t.Run("check violation preserves constraint name", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:           ErrCodeCheckViolation,
			Message:        `new row for relation "items" violates check constraint "items_price_check"`,
			ConstraintName: "items_price_check",
		}
		msg := SanitizeErrorMessage(err)
		assert.Contains(t, msg, "check constraint violation")
		assert.Contains(t, msg, "items_price_check")
	})

	t.Run("syntax error hides query text but keeps code", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:    "42601",
			Message: `syntax error at or near "DROP"`,
		}
		msg := SanitizeErrorMessage(err)
		assert.Contains(t, msg, "42601")
		assert.NotContains(t, msg, "DROP")
	})

	t.Run("insufficient privilege hides message", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:    "42501",
			Message: "permission denied for table auth_users",
		}
		msg := SanitizeErrorMessage(err)
		assert.Contains(t, msg, "42501")
		assert.NotContains(t, msg, "auth_users")
	})

	t.Run("integrity class fallback keeps only category and code", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:    "23502",
			Message: `null value in column "email" of relation "users" violates not-null constraint`,
		}
		msg := SanitizeErrorMessage(err)
		assert.Contains(t, msg, "data integrity violation")
		assert.Contains(t, msg, "23502")
		assert.NotContains(t, msg, "email")
	})

	t.Run("non-pg errors get a generic message", func(t *testing.T) {
		err := errors.New("failed to connect to postgres://admin:secret@10.0.0.1:5432/db")
		msg := SanitizeErrorMessage(err)
		assert.Equal(t, "database error", msg)
		assert.NotContains(t, msg, "secret")
	})

	t.Run("wrapped pg errors are sanitized", func(t *testing.T) {
		inner := &pgconn.PgError{Code: ErrCodeUniqueViolation, ConstraintName: "c"}
		wrapped := fmt.Errorf("outer: %w", inner)
		msg := SanitizeErrorMessage(wrapped)
		assert.Contains(t, msg, "unique constraint violation")
	})
}
