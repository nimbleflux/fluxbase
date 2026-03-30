package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTablesList_Success(t *testing.T) {
	resetTableFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/tables")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"schema": "public", "name": "users", "type": "table"},
			{"schema": "public", "name": "products", "type": "table"},
			{"schema": "auth", "name": "users", "type": "table"},
		})
	})
	defer cleanup()

	err := runTablesList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 3)
	assert.Equal(t, "users", result[0]["name"])
	assert.Equal(t, "products", result[1]["name"])
}

func TestTablesList_Empty(t *testing.T) {
	resetTableFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	// Empty list prints "No tables found." via fmt.Println (not to formatter)
	err := runTablesList(nil, []string{})
	require.NoError(t, err)
}

func TestTablesList_WithSchema(t *testing.T) {
	resetTableFlags()
	tableSchema = "auth"
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "auth", r.URL.Query().Get("schema"))
		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"schema": "auth", "name": "users", "type": "table"},
		})
	})
	defer cleanup()

	err := runTablesList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 1)
}

func TestTablesList_APIError(t *testing.T) {
	resetTableFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runTablesList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestTablesDescribe_Success(t *testing.T) {
	resetTableFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/tables/public/users/columns")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"name": "id", "data_type": "integer", "is_nullable": false, "default": "nextval('users_id_seq'::regclass)"},
			{"name": "email", "data_type": "text", "is_nullable": false, "default": ""},
		})
	})
	defer cleanup()

	err := runTablesDescribe(nil, []string{"users"})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "id", result[0]["name"])
	assert.Equal(t, "email", result[1]["name"])
}

func TestTablesDescribe_WithSchema(t *testing.T) {
	resetTableFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/api/v1/admin/tables/auth/users/columns")
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runTablesDescribe(nil, []string{"auth.users"})
	require.NoError(t, err)
}

func TestTablesQuery_Success(t *testing.T) {
	resetTableFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/tables/public/products")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"id": float64(1), "name": "Widget", "price": float64(9.99)},
			{"id": float64(2), "name": "Gadget", "price": float64(19.99)},
		})
	})
	defer cleanup()

	err := runTablesQuery(nil, []string{"products"})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
}

func TestTablesQuery_Empty(t *testing.T) {
	resetTableFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	// Empty result prints "No records found." via fmt.Println
	err := runTablesQuery(nil, []string{"products"})
	require.NoError(t, err)
}

func TestTablesQuery_WithFilters(t *testing.T) {
	resetTableFlags()
	tableLimit = 10
	tableOffset = 5
	tableOrderBy = "name.asc"
	tableWhere = "price=eq.10"

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "10", q.Get("limit"))
		assert.Equal(t, "5", q.Get("offset"))
		assert.Equal(t, "name.asc", q.Get("order"))
		assert.Equal(t, "eq.10", q.Get("price"))

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"id": float64(1), "name": "test"},
		})
	})
	defer cleanup()

	err := runTablesQuery(nil, []string{"products"})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 1)
}

func TestTablesInsert_Success(t *testing.T) {
	resetTableFlags()
	tableData = `{"name":"Test Product","price":29.99}`

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/tables/public/products")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "Test Product", body["name"])

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id":    float64(1),
			"name":  "Test Product",
			"price": float64(29.99),
		})
	})
	defer cleanup()

	err := runTablesInsert(nil, []string{"products"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "Test Product", result["name"])
}

func TestTablesInsert_NoData(t *testing.T) {
	resetTableFlags()
	tableData = ""
	tableFile = ""

	err := runTablesInsert(nil, []string{"products"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "either --data or --file is required")
}

func TestTablesInsert_InvalidJSON(t *testing.T) {
	resetTableFlags()
	tableData = `{invalid json`

	err := runTablesInsert(nil, []string{"products"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON data")
}

func TestTablesUpdate_Success(t *testing.T) {
	resetTableFlags()
	tableData = `{"name":"Updated"}`
	tableWhere = "id=eq.1"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/tables/public/products")
		assert.Equal(t, "eq.1", r.URL.Query().Get("id"))

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	// Success prints via fmt.Println, just verify no error
	err := runTablesUpdate(nil, []string{"products"})
	require.NoError(t, err)
}

func TestTablesDelete_Success(t *testing.T) {
	resetTableFlags()
	tableWhere = "id=eq.1"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/tables/public/products")
		assert.Equal(t, "eq.1", r.URL.Query().Get("id"))

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	// Success prints via fmt.Println, just verify no error
	err := runTablesDelete(nil, []string{"products"})
	require.NoError(t, err)
}

func TestTablesDelete_APIError(t *testing.T) {
	resetTableFlags()
	tableWhere = "id=eq.999"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "record not found")
	})
	defer cleanup()

	err := runTablesDelete(nil, []string{"products"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}
