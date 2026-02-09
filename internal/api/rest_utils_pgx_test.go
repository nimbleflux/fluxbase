package api

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// pgxRowsToJSON Tests
// =============================================================================

// mockRows is a simple mock for pgx.Rows for testing pgxRowsToJSON
type mockRows struct {
	fields    []pgconn.FieldDescription
	rows      [][]interface{}
	nextIndex int
	nextErr   error
	scanErr   error
}

func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription {
	return m.fields
}

func (m *mockRows) Next() bool {
	if m.nextErr != nil {
		return false
	}
	if m.nextIndex >= len(m.rows) {
		return false
	}
	m.nextIndex++
	return m.nextIndex <= len(m.rows)
}

func (m *mockRows) Scan(dest ...interface{}) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	if m.nextIndex > len(m.rows) {
		return pgx.ErrNoRows
	}
	row := m.rows[m.nextIndex-1]
	for i := range dest {
		if i < len(row) {
			if ptr, ok := dest[i].(*interface{}); ok {
				*ptr = row[i]
			}
		}
	}
	return nil
}

func (m *mockRows) Err() error {
	return m.nextErr
}

func (m *mockRows) Close() {}

func (m *mockRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (m *mockRows) Values() ([]interface{}, error) {
	if m.nextIndex > len(m.rows) {
		return nil, pgx.ErrNoRows
	}
	return m.rows[m.nextIndex-1], nil
}

func (m *mockRows) RawValues() [][]byte {
	return nil
}

func (m *mockRows) Conn() *pgx.Conn {
	return nil
}

func TestPgxRowsToJSON_SimpleTypes(t *testing.T) {
	tests := []struct {
		name     string
		fields   []pgconn.FieldDescription
		rows     [][]interface{}
		expected []map[string]interface{}
	}{
		{
			name: "single row with string and int",
			fields: []pgconn.FieldDescription{
				{Name: "name"},
				{Name: "age"},
			},
			rows: [][]interface{}{
				{"John", 30},
			},
			expected: []map[string]interface{}{
				{"name": "John", "age": 30},
			},
		},
		{
			name: "multiple rows",
			fields: []pgconn.FieldDescription{
				{Name: "id"},
				{Name: "email"},
			},
			rows: [][]interface{}{
				{1, "user1@example.com"},
				{2, "user2@example.com"},
				{3, "user3@example.com"},
			},
			expected: []map[string]interface{}{
				{"id": 1, "email": "user1@example.com"},
				{"id": 2, "email": "user2@example.com"},
				{"id": 3, "email": "user3@example.com"},
			},
		},
		{
			name: "boolean values",
			fields: []pgconn.FieldDescription{
				{Name: "active"},
				{Name: "verified"},
			},
			rows: [][]interface{}{
				{true, false},
				{false, true},
			},
			expected: []map[string]interface{}{
				{"active": true, "verified": false},
				{"active": false, "verified": true},
			},
		},
		{
			name: "null values",
			fields: []pgconn.FieldDescription{
				{Name: "name"},
				{Name: "middle_name"},
				{Name: "age"},
			},
			rows: [][]interface{}{
				{"John", nil, 30},
			},
			expected: []map[string]interface{}{
				{"name": "John", "middle_name": nil, "age": 30},
			},
		},
		{
			name: "float values",
			fields: []pgconn.FieldDescription{
				{Name: "price"},
				{Name: "tax_rate"},
			},
			rows: [][]interface{}{
				{19.99, 0.08},
				{29.99, 0.10},
			},
			expected: []map[string]interface{}{
				{"price": 19.99, "tax_rate": 0.08},
				{"price": 29.99, "tax_rate": 0.10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRows{
				fields: tt.fields,
				rows:   tt.rows,
			}

			result, err := pgxRowsToJSON(mock)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPgxRowsToJSON_ByteArrays(t *testing.T) {
	t.Run("JSON byte array is unmarshaled", func(t *testing.T) {
		jsonData := `{"key":"value","nested":{"number":123}}`
		fields := []pgconn.FieldDescription{
			{Name: "id"},
			{Name: "metadata"},
		}
		rows := [][]interface{}{
			{1, []byte(jsonData)},
		}

		mock := &mockRows{fields: fields, rows: rows}
		result, err := pgxRowsToJSON(mock)
		require.NoError(t, err)
		require.Len(t, result, 1)

		// metadata should be unmarshaled as a map
		metadata, ok := result[0]["metadata"].(map[string]interface{})
		require.True(t, ok, "metadata should be unmarshaled as map")
		assert.Equal(t, "value", metadata["key"])

		nested, ok := metadata["nested"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(123), nested["number"])
	})

	t.Run("invalid JSON byte array is treated as string", func(t *testing.T) {
		invalidJSON := `{not valid json}`
		fields := []pgconn.FieldDescription{
			{Name: "id"},
			{Name: "data"},
		}
		rows := [][]interface{}{
			{1, []byte(invalidJSON)},
		}

		mock := &mockRows{fields: fields, rows: rows}
		result, err := pgxRowsToJSON(mock)
		require.NoError(t, err)
		require.Len(t, result, 1)

		// Invalid JSON should be returned as string
		data, ok := result[0]["data"].(string)
		require.True(t, ok, "data should be string for invalid JSON")
		assert.Equal(t, invalidJSON, data)
	})

	t.Run("plain text byte array is treated as string", func(t *testing.T) {
		textData := "hello world"
		fields := []pgconn.FieldDescription{
			{Name: "id"},
			{Name: "message"},
		}
		rows := [][]interface{}{
			{1, []byte(textData)},
		}

		mock := &mockRows{fields: fields, rows: rows}
		result, err := pgxRowsToJSON(mock)
		require.NoError(t, err)
		require.Len(t, result, 1)

		message, ok := result[0]["message"].(string)
		require.True(t, ok, "message should be string")
		assert.Equal(t, textData, message)
	})
}

func TestPgxRowsToJSON_UUIDBytes(t *testing.T) {
	t.Run("UUID byte array is converted to string", func(t *testing.T) {
		uid := uuid.New()
		fields := []pgconn.FieldDescription{
			{Name: "id"},
			{Name: "user_id"},
		}
		rows := [][]interface{}{
			{1, [16]byte(uid)},
		}

		mock := &mockRows{fields: fields, rows: rows}
		result, err := pgxRowsToJSON(mock)
		require.NoError(t, err)
		require.Len(t, result, 1)

		userID, ok := result[0]["user_id"].(string)
		require.True(t, ok, "user_id should be string")
		assert.Equal(t, uid.String(), userID)
	})
}

func TestPgxRowsToJSON_EmptyResults(t *testing.T) {
	t.Run("no rows", func(t *testing.T) {
		fields := []pgconn.FieldDescription{
			{Name: "id"},
			{Name: "name"},
		}

		mock := &mockRows{
			fields: fields,
			rows:   [][]interface{}{},
		}

		result, err := pgxRowsToJSON(mock)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("single empty row", func(t *testing.T) {
		fields := []pgconn.FieldDescription{
			{Name: "id"},
		}
		rows := [][]interface{}{
			{nil},
		}

		mock := &mockRows{fields: fields, rows: rows}
		result, err := pgxRowsToJSON(mock)
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Nil(t, result[0]["id"])
	})
}

func TestPgxRowsToJSON_Errors(t *testing.T) {
	t.Run("scan error propagates", func(t *testing.T) {
		fields := []pgconn.FieldDescription{
			{Name: "id"},
		}
		rows := [][]interface{}{
			{1},
		}

		mock := &mockRows{
			fields:  fields,
			rows:    rows,
			scanErr: assert.AnError,
		}

		result, err := pgxRowsToJSON(mock)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("rows.Err() propagates", func(t *testing.T) {
		fields := []pgconn.FieldDescription{
			{Name: "id"},
		}
		rows := [][]interface{}{
			{1},
			{2},
		}

		mock := &mockRows{
			fields:  fields,
			rows:    rows,
			nextErr: assert.AnError,
		}

		result, err := pgxRowsToJSON(mock)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestPgxRowsToJSON_MixedTypes(t *testing.T) {
	uid := uuid.New()
	jsonData := `{"setting":"value"}`
	fields := []pgconn.FieldDescription{
		{Name: "id"},
		{Name: "name"},
		{Name: "active"},
		{Name: "count"},
		{Name: "rating"},
		{Name: "uuid_col"},
		{Name: "json_col"},
	}
	rows := [][]interface{}{
		{1, "Test", true, int32(100), float64(4.5), [16]byte(uid), []byte(jsonData)},
	}

	mock := &mockRows{fields: fields, rows: rows}
	result, err := pgxRowsToJSON(mock)
	require.NoError(t, err)
	require.Len(t, result, 1)

	row := result[0]
	assert.Equal(t, 1, row["id"])
	assert.Equal(t, "Test", row["name"])
	assert.Equal(t, true, row["active"])
	assert.Equal(t, int32(100), row["count"])
	assert.Equal(t, 4.5, row["rating"])
	assert.Equal(t, uid.String(), row["uuid_col"])

	jsonCol, ok := row["json_col"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value", jsonCol["setting"])
}
