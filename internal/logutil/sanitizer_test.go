package logutil

import (
	"testing"
)

func TestSanitizeSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string literal",
			input:    "SELECT * FROM users WHERE name = 'John'",
			expected: "SELECT * FROM users WHERE name = '<redacted>'",
		},
		{
			name:     "numeric literal",
			input:    "SELECT * FROM users WHERE id = 123",
			expected: "SELECT * FROM users WHERE id = <num>",
		},
		{
			name:     "boolean literal",
			input:    "UPDATE users SET active = TRUE WHERE id = 1",
			expected: "UPDATE users SET active = <bool> WHERE id = <num>",
		},
		{
			name:     "NULL literal",
			input:    "UPDATE users SET deleted_at = NULL WHERE id = 5",
			expected: "UPDATE users SET deleted_at = <null> WHERE id = <num>",
		},
		{
			name:     "complex query with multiple literals",
			input:    "SELECT * FROM users WHERE email = 'test@example.com' AND age > 25 AND active = FALSE",
			expected: "SELECT * FROM users WHERE email = '<redacted>' AND age > <num> AND active = <bool>",
		},
		{
			name:     "escaped quotes in string",
			input:    "SELECT * FROM users WHERE name = 'O''Reilly'",
			expected: "SELECT * FROM users WHERE name = '<redacted>'",
		},
		{
			name:     "IPv4 address",
			input:    "INSERT INTO logs (ip) VALUES ('192.168.1.1')",
			expected: "INSERT INTO logs (ip) VALUES ('<redacted>')",
		},
		{
			name:     "UUID in query",
			input:    "SELECT * FROM users WHERE id = '550e8400-e29b-41d4-a716-446655440000'",
			expected: "SELECT * FROM users WHERE id = '<redacted>'",
		},
		{
			name:     "float number",
			input:    "SELECT * FROM products WHERE price > 99.99",
			expected: "SELECT * FROM products WHERE price > <num>",
		},
		{
			name:     "scientific notation",
			input:    "SELECT * FROM measurements WHERE value > 1.5e10",
			expected: "SELECT * FROM measurements WHERE value > <num>",
		},
		{
			name:     "parameter placeholders preserved",
			input:    "SELECT * FROM users WHERE id = $1 AND name = $2",
			expected: "SELECT * FROM users WHERE id = $1 AND name = $2",
		},
		{
			name:     "INSERT with values",
			input:    "INSERT INTO users (name, email, age) VALUES ('John', 'john@example.com', 30)",
			expected: "INSERT INTO users (name, email, age) VALUES ('<redacted>', '<redacted>', <num>)",
		},
		{
			name:     "UPDATE with SET clause",
			input:    "UPDATE users SET name = 'Jane', age = 25 WHERE id = 123",
			expected: "UPDATE users SET name = '<redacted>', age = <num> WHERE id = <num>",
		},
		{
			name:     "dollar-quoted string",
			input:    "$$This is a dollar-quoted string$$",
			expected: "$$<redacted>$$",
		},
		{
			name:     "dollar-tagged string",
			input:    "$function$CREATE FUNCTION test() RETURNS int$function$",
			expected: "$<redacted>$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeSQL(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeSQL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractDDLMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CREATE TABLE",
			input:    "CREATE TABLE users (id SERIAL, name TEXT)",
			expected: "CREATE TABLE users",
		},
		{
			name:     "CREATE TABLE IF NOT EXISTS",
			input:    "CREATE TABLE IF NOT EXISTS users (id SERIAL)",
			expected: "CREATE TABLE users",
		},
		{
			name:     "CREATE INDEX",
			input:    "CREATE INDEX idx_users_email ON users(email)",
			expected: "CREATE INDEX idx_users_email",
		},
		{
			name:     "CREATE UNIQUE INDEX",
			input:    "CREATE UNIQUE INDEX idx_users_email ON users(email)",
			expected: "CREATE INDEX idx_users_email",
		},
		{
			name:     "CREATE OR REPLACE FUNCTION",
			input:    "CREATE OR REPLACE FUNCTION get_user(id INTEGER) RETURNS TABLE",
			expected: "CREATE FUNCTION get_user",
		},
		{
			name:     "CREATE MATERIALIZED VIEW",
			input:    "CREATE MATERIALIZED VIEW user_stats AS SELECT COUNT(*) FROM users",
			expected: "CREATE VIEW user_stats",
		},
		{
			name:     "CREATE VIEW",
			input:    "CREATE VIEW active_users AS SELECT * FROM users WHERE active = TRUE",
			expected: "CREATE VIEW active_users",
		},
		{
			name:     "CREATE TRIGGER",
			input:    "CREATE TRIGGER update_timestamp BEFORE UPDATE ON users",
			expected: "CREATE TRIGGER update_timestamp",
		},
		{
			name:     "CREATE SCHEMA",
			input:    "CREATE SCHEMA app_data",
			expected: "CREATE SCHEMA app_data",
		},
		{
			name:     "ALTER TABLE ADD COLUMN",
			input:    "ALTER TABLE users ADD COLUMN email TEXT",
			expected: "ALTER TABLE users ADD",
		},
		{
			name:     "ALTER TABLE DROP COLUMN",
			input:    "ALTER TABLE users DROP COLUMN old_field",
			expected: "ALTER TABLE users DROP",
		},
		{
			name:     "ALTER TABLE RENAME COLUMN",
			input:    "ALTER TABLE users RENAME COLUMN name TO full_name",
			expected: "ALTER TABLE users RENAME COLUMN",
		},
		{
			name:     "ALTER TABLE ADD CONSTRAINT",
			input:    "ALTER TABLE users ADD CONSTRAINT pk_users PRIMARY KEY (id)",
			expected: "ALTER TABLE users ADD CONSTRAINT",
		},
		{
			name:     "DROP TABLE",
			input:    "DROP TABLE users",
			expected: "DROP TABLE users",
		},
		{
			name:     "DROP TABLE IF EXISTS",
			input:    "DROP TABLE IF EXISTS users",
			expected: "DROP TABLE users",
		},
		{
			name:     "DROP INDEX CONCURRENTLY",
			input:    "DROP INDEX CONCURRENTLY idx_users_email",
			expected: "DROP INDEX idx_users_email",
		},
		{
			name:     "DROP FUNCTION",
			input:    "DROP FUNCTION IF EXISTS get_user INTEGER",
			expected: "DROP FUNCTION get_user",
		},
		{
			name:     "TRUNCATE TABLE",
			input:    "TRUNCATE TABLE users",
			expected: "TRUNCATE TABLE users",
		},
		{
			name:     "TRUNCATE TABLE ONLY",
			input:    "TRUNCATE ONLY users RESTART IDENTITY",
			expected: "TRUNCATE TABLE users",
		},
		{
			name:     "RENAME TABLE",
			input:    "RENAME TABLE old_users TO users",
			expected: "RENAME TABLE old_users",
		},
		{
			name:     "GRANT SELECT",
			input:    "GRANT SELECT ON TABLE users TO app_user",
			expected: "GRANT ON TABLE users",
		},
		{
			name:     "REVOKE INSERT",
			input:    "REVOKE INSERT ON TABLE users FROM app_user",
			expected: "REVOKE ON TABLE users",
		},
		{
			name:     "COMMENT ON TABLE",
			input:    "COMMENT ON TABLE users IS 'User information'",
			expected: "COMMENT ON TABLE users",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   \n\t  ",
			expected: "",
		},
		{
			name:     "CREATE TYPE",
			input:    "CREATE TYPE user_status AS ENUM ('active', 'inactive')",
			expected: "CREATE TYPE user_status",
		},
		{
			name:     "CREATE EXTENSION",
			input:    "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"",
			expected: "CREATE EXTENSION uuid-ossp",
		},
		{
			name:     "CREATE FUNCTION with parameters",
			input:    "CREATE OR REPLACE FUNCTION get_user(id INTEGER) RETURNS TABLE",
			expected: "CREATE FUNCTION get_user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ExtractDDLMetadata(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractDDLMetadata() = %q, want %q", result, tt.expected)
			}
		})
	}
}
