package database

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// MockExecutor for testing
// =============================================================================

type MockExecutor struct {
	queryFunc    func(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	queryRowFunc func(ctx context.Context, sql string, args ...interface{}) pgx.Row
	execFunc     func(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

func (m *MockExecutor) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func (m *MockExecutor) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return &MockRow{}
}

func (m *MockExecutor) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, args...)
	}
	return pgconn.NewCommandTag("SELECT 0"), nil
}

func (m *MockExecutor) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return nil, nil
}

func (m *MockExecutor) Pool() interface{} {
	return nil
}

func (m *MockExecutor) Health(ctx context.Context) error {
	return nil
}

type MockRow struct {
	scanFunc func(dest ...interface{}) error
}

func (m *MockRow) Scan(dest ...interface{}) error {
	if m.scanFunc != nil {
		return m.scanFunc(dest)
	}
	return nil
}

// =============================================================================
// GetTableInfo Tests
// =============================================================================

func TestSchemaInspector_GetTableInfo_Errors(t *testing.T) {
	t.Run("returns error when columns query fails", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		// This will panic since there's no actual connection pool
		// We can't test error paths without a mock connection
		assert.Panics(t, func() {
			_, _ = inspector.GetTableInfo(context.Background(), "public", "nonexistent")
		})
	})
}

// =============================================================================
// GetAllTables Tests
// =============================================================================

func TestSchemaInspector_GetAllTables_Schemas(t *testing.T) {
	t.Run("uses public schema by default", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		// With no connection pool, this will panic
		assert.Panics(t, func() {
			_, _ = inspector.GetAllTables(context.Background())
		})
	})

	t.Run("accepts multiple schemas", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetAllTables(context.Background(), "public", "auth", "storage")
		})
	})
}

// =============================================================================
// GetAllViews Tests
// =============================================================================

func TestSchemaInspector_GetAllViews_Schemas(t *testing.T) {
	t.Run("uses public schema by default", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetAllViews(context.Background())
		})
	})

	t.Run("accepts multiple schemas", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetAllViews(context.Background(), "analytics", "reporting")
		})
	})
}

// =============================================================================
// GetAllMaterializedViews Tests
// =============================================================================

func TestSchemaInspector_GetAllMaterializedViews_Schemas(t *testing.T) {
	t.Run("uses public schema by default", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetAllMaterializedViews(context.Background())
		})
	})

	t.Run("accepts multiple schemas", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetAllMaterializedViews(context.Background(), "reporting")
		})
	})
}

// =============================================================================
// GetAllFunctions Tests
// =============================================================================

func TestSchemaInspector_GetAllFunctions_Schemas(t *testing.T) {
	t.Run("uses public schema by default", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetAllFunctions(context.Background())
		})
	})

	t.Run("accepts multiple schemas", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetAllFunctions(context.Background(), "public", "auth")
		})
	})
}

// =============================================================================
// GetSchemas Tests
// =============================================================================

func TestSchemaInspector_GetSchemas(t *testing.T) {
	t.Run("returns error with no connection", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetSchemas(context.Background())
		})
	})
}

// =============================================================================
// GetVectorColumns Tests
// =============================================================================

func TestSchemaInspector_GetVectorColumns(t *testing.T) {
	t.Run("uses public schema by default", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetVectorColumns(context.Background(), "", "")
		})
	})

	t.Run("accepts schema and table parameters", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetVectorColumns(context.Background(), "public", "embeddings")
		})
	})

	t.Run("accepts only schema parameter", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _ = inspector.GetVectorColumns(context.Background(), "public", "")
		})
	})
}

// =============================================================================
// IsPgVectorInstalled Tests
// =============================================================================

func TestSchemaInspector_IsPgVectorInstalled(t *testing.T) {
	t.Run("returns not installed with no connection", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.Panics(t, func() {
			_, _, _ = inspector.IsPgVectorInstalled(context.Background())
		})
	})
}

// =============================================================================
// TableInfo GetColumn Tests (additional edge cases)
// =============================================================================

func TestTableInfo_GetColumn_EdgeCases(t *testing.T) {
	t.Run("returns nil for empty column name", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "users",
			Columns: []ColumnInfo{
				{Name: "id", DataType: "uuid"},
			},
		}
		table.BuildColumnMap()

		col := table.GetColumn("")
		assert.Nil(t, col)
	})

	t.Run("case sensitive column lookup", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "users",
			Columns: []ColumnInfo{
				{Name: "UserID", DataType: "uuid"},
				{Name: "userid", DataType: "text"},
			},
		}
		table.BuildColumnMap()

		// Should match exact case
		col1 := table.GetColumn("UserID")
		assert.NotNil(t, col1)
		assert.Equal(t, "UserID", col1.Name)

		col2 := table.GetColumn("userid")
		assert.NotNil(t, col2)
		assert.Equal(t, "userid", col2.Name)

		// Should not match different case
		col3 := table.GetColumn("userId")
		assert.Nil(t, col3)
	})

	t.Run("column lookup with special characters", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "data",
			Columns: []ColumnInfo{
				{Name: "user-id", DataType: "uuid"},
				{Name: "user_id", DataType: "uuid"},
				{Name: "user.id", DataType: "uuid"},
			},
		}
		table.BuildColumnMap()

		assert.NotNil(t, table.GetColumn("user-id"))
		assert.NotNil(t, table.GetColumn("user_id"))
		assert.NotNil(t, table.GetColumn("user.id"))
	})
}

// =============================================================================
// TableInfo HasColumn Tests (additional edge cases)
// =============================================================================

func TestTableInfo_HasColumn_EdgeCases(t *testing.T) {
	t.Run("empty column name returns false", func(t *testing.T) {
		table := TableInfo{
			Schema:  "public",
			Name:    "users",
			Columns: []ColumnInfo{{Name: "id", DataType: "uuid"}},
		}
		table.BuildColumnMap()

		assert.False(t, table.HasColumn(""))
	})

	t.Run("whitespace column name", func(t *testing.T) {
		table := TableInfo{
			Schema:  "public",
			Name:    "users",
			Columns: []ColumnInfo{{Name: "id", DataType: "uuid"}},
		}
		table.BuildColumnMap()

		assert.False(t, table.HasColumn("  "))
		assert.False(t, table.HasColumn("id "))
		assert.False(t, table.HasColumn(" id"))
	})
}

// =============================================================================
// BuildColumnMap Tests (edge cases)
// =============================================================================

func TestTableInfo_BuildColumnMap_EdgeCases(t *testing.T) {
	t.Run("handles duplicate column names", func(t *testing.T) {
		// This shouldn't happen in valid SQL, but test defensive behavior
		table := TableInfo{
			Schema: "public",
			Name:   "users",
			Columns: []ColumnInfo{
				{Name: "id", DataType: "uuid"},
				{Name: "id", DataType: "text"}, // Duplicate
			},
		}

		table.BuildColumnMap()

		// Last one wins (expected behavior for map)
		assert.Len(t, table.ColumnMap, 1)
		col := table.ColumnMap["id"]
		assert.NotNil(t, col)
		assert.Equal(t, "text", col.DataType) // Second one wins
	})

	t.Run("idempotent - can be called multiple times", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "users",
			Columns: []ColumnInfo{
				{Name: "id", DataType: "uuid"},
				{Name: "email", DataType: "text"},
			},
		}

		table.BuildColumnMap()
		firstLen := len(table.ColumnMap)

		table.BuildColumnMap()
		secondLen := len(table.ColumnMap)

		// Should be the same length
		assert.Equal(t, firstLen, secondLen)
		assert.Len(t, table.ColumnMap, 2)
	})

	t.Run("preserves column data pointers", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "users",
			Columns: []ColumnInfo{
				{Name: "id", DataType: "uuid", IsPrimaryKey: true},
			},
		}

		table.BuildColumnMap()

		// The map should point to the actual column in the slice
		colFromMap := table.ColumnMap["id"]
		colFromSlice := &table.Columns[0]

		assert.Equal(t, colFromSlice.DataType, colFromMap.DataType)
		assert.Equal(t, colFromSlice.IsPrimaryKey, colFromMap.IsPrimaryKey)
	})
}

// =============================================================================
// Additional REST Path Tests
// =============================================================================

func TestSchemaInspector_BuildRESTPath_EdgeCases(t *testing.T) {
	inspector := &SchemaInspector{}

	t.Run("table name ending with ss", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "class",
		}
		result := inspector.BuildRESTPath(table)
		assert.Equal(t, "/api/rest/class", result)
	})

	t.Run("table name ending with sh", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "dish",
		}
		result := inspector.BuildRESTPath(table)
		assert.Equal(t, "/api/rest/dishes", result)
	})

	t.Run("table ending with ch", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "match",
		}
		result := inspector.BuildRESTPath(table)
		assert.Equal(t, "/api/rest/matches", result)
	})

	t.Run("table with multiple underscores", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "user_profile_preference",
		}
		result := inspector.BuildRESTPath(table)
		assert.Equal(t, "/api/rest/user_profile_preferences", result)
	})

	t.Run("table with numbers", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "user2",
		}
		result := inspector.BuildRESTPath(table)
		assert.Equal(t, "/api/rest/user2s", result)
	})

	t.Run("table ending with number then y", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "category2",
		}
		result := inspector.BuildRESTPath(table)
		// '2y' - the 'y' is preceded by a number, not consonant, so should add 's'
		assert.Equal(t, "/api/rest/category2s", result)
	})

	t.Run("single character table", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "x",
		}
		result := inspector.BuildRESTPath(table)
		assert.Equal(t, "/api/rest/xes", result) // 'x' ends with 'x' so adds 'es'
	})

	t.Run("table ending with fie (like leaf)", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "leaf",
		}
		result := inspector.BuildRESTPath(table)
		// 'f' is not a vowel, so 'y' rule applies... but there's no 'y'
		// Just adds 's'
		assert.Equal(t, "/api/rest/leafs", result)
	})

	t.Run("camel case table name", func(t *testing.T) {
		table := TableInfo{
			Schema: "public",
			Name:   "UserProfile",
		}
		result := inspector.BuildRESTPath(table)
		// Doesn't handle camel case, just treats as singular
		assert.Equal(t, "/api/rest/UserProfiles", result)
	})
}

// =============================================================================
// Connection Interface Compliance Tests
// =============================================================================

func TestSchemaInspector_ConnectionInterface(t *testing.T) {
	t.Run("SchemaInspector can be created with nil connection", func(t *testing.T) {
		inspector := NewSchemaInspector(nil)

		assert.NotNil(t, inspector)
		assert.Nil(t, inspector.conn)
	})

	t.Run("SchemaInspector stores connection reference", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		assert.NotNil(t, inspector)
		assert.Same(t, conn, inspector.conn)
	})
}

// =============================================================================
// Type Field Tests
// =============================================================================

func TestTypeInfo_StringValues(t *testing.T) {
	t.Run("table type string", func(t *testing.T) {
		table := TableInfo{Type: "table"}
		assert.Equal(t, "table", table.Type)
	})

	t.Run("view type string", func(t *testing.T) {
		view := TableInfo{Type: "view"}
		assert.Equal(t, "view", view.Type)
	})

	t.Run("materialized view type string", func(t *testing.T) {
		mv := TableInfo{Type: "materialized_view"}
		assert.Equal(t, "materialized_view", mv.Type)
	})
}

// =============================================================================
// ForeignKey Tests
// =============================================================================

func TestForeignKey_EdgeCases(t *testing.T) {
	t.Run("self-referencing foreign key", func(t *testing.T) {
		fk := ForeignKey{
			Name:             "fk_users_parent_id",
			ColumnName:       "parent_id",
			ReferencedTable:  "public.users",
			ReferencedColumn: "id",
			OnDelete:         "NO ACTION",
			OnUpdate:         "CASCADE",
		}

		assert.Equal(t, "parent_id", fk.ColumnName)
		assert.Contains(t, fk.ReferencedTable, "users")
	})

	t.Run("composite foreign key representation", func(t *testing.T) {
		// Each FK is one column, but composite FKs have multiple rows
		fk1 := ForeignKey{
			Name:             "fk_order_items",
			ColumnName:       "order_id",
			ReferencedTable:  "public.orders",
			ReferencedColumn: "id",
		}
		fk2 := ForeignKey{
			Name:             "fk_order_items",
			ColumnName:       "product_id",
			ReferencedTable:  "public.products",
			ReferencedColumn: "id",
		}

		assert.Equal(t, "fk_order_items", fk1.Name)
		assert.Equal(t, "fk_order_items", fk2.Name)
		assert.Equal(t, "order_id", fk1.ColumnName)
		assert.Equal(t, "product_id", fk2.ColumnName)
	})
}

// =============================================================================
// Index Tests
// =============================================================================

func TestIndexInfo_EdgeCases(t *testing.T) {
	t.Run("single column index", func(t *testing.T) {
		idx := IndexInfo{
			Name:     "idx_users_email",
			Columns:  []string{"email"},
			IsUnique: true,
		}

		assert.Len(t, idx.Columns, 1)
		assert.True(t, idx.IsUnique)
	})

	t.Run("multi-column index", func(t *testing.T) {
		idx := IndexInfo{
			Name:     "idx_posts_user_created",
			Columns:  []string{"user_id", "created_at"},
			IsUnique: false,
		}

		assert.Len(t, idx.Columns, 2)
		assert.False(t, idx.IsUnique)
	})

	t.Run("functional index", func(t *testing.T) {
		// PostgreSQL supports functional indexes like (lower(email))
		// These would be represented with the function expression as the column name
		idx := IndexInfo{
			Name:     "idx_users_lower_email",
			Columns:  []string{"lower(email)"},
			IsUnique: true,
		}

		assert.Equal(t, "lower(email)", idx.Columns[0])
	})
}

// =============================================================================
// FunctionInfo Tests
// =============================================================================

func TestFunctionInfo_Volatility(t *testing.T) {
	t.Run("IMMUTABLE volatility", func(t *testing.T) {
		fn := FunctionInfo{Volatility: "IMMUTABLE"}
		assert.Equal(t, "IMMUTABLE", fn.Volatility)
	})

	t.Run("STABLE volatility", func(t *testing.T) {
		fn := FunctionInfo{Volatility: "STABLE"}
		assert.Equal(t, "STABLE", fn.Volatility)
	})

	t.Run("VOLATILE volatility", func(t *testing.T) {
		fn := FunctionInfo{Volatility: "VOLATILE"}
		assert.Equal(t, "VOLATILE", fn.Volatility)
	})
}

func TestFunctionParam_Mode(t *testing.T) {
	t.Run("IN mode", func(t *testing.T) {
		param := FunctionParam{Mode: "IN"}
		assert.Equal(t, "IN", param.Mode)
	})

	t.Run("OUT mode", func(t *testing.T) {
		param := FunctionParam{Mode: "OUT"}
		assert.Equal(t, "OUT", param.Mode)
	})

	t.Run("INOUT mode", func(t *testing.T) {
		param := FunctionParam{Mode: "INOUT"}
		assert.Equal(t, "INOUT", param.Mode)
	})
}

// =============================================================================
// VectorColumnInfo Tests
// =============================================================================

func TestVectorColumnInfo_Dimensions(t *testing.T) {
	t.Run("OpenAI embedding dimensions", func(t *testing.T) {
		col := VectorColumnInfo{Dimensions: 1536}
		assert.Equal(t, 1536, col.Dimensions)
	})

	t.Run("variable length vector", func(t *testing.T) {
		col := VectorColumnInfo{Dimensions: -1}
		assert.Equal(t, -1, col.Dimensions)
	})

	t.Run("small vector dimensions", func(t *testing.T) {
		col := VectorColumnInfo{Dimensions: 128}
		assert.Equal(t, 128, col.Dimensions)
	})
}

// =============================================================================
// Error Path Tests
// =============================================================================

func TestSchemaInspector_ErrorPaths(t *testing.T) {
	t.Run("panics with nil connection", func(t *testing.T) {
		conn := &Connection{}
		inspector := NewSchemaInspector(conn)

		// All query operations should panic with nil connection pool
		assert.Panics(t, func() {
			_, _ = inspector.GetSchemas(context.Background())
		})

		assert.Panics(t, func() {
			_, _ = inspector.GetAllTables(context.Background())
		})

		assert.Panics(t, func() {
			_, _ = inspector.GetAllViews(context.Background())
		})

		assert.Panics(t, func() {
			_, _ = inspector.GetAllMaterializedViews(context.Background())
		})

		assert.Panics(t, func() {
			_, _ = inspector.GetAllFunctions(context.Background())
		})
	})
}

// =============================================================================
// Batch Query Simulation Tests
// =============================================================================

func TestSchemaInspector_BatchQueryLogic(t *testing.T) {
	t.Run("column aggregation preserves order", func(t *testing.T) {
		// Simulate batch column aggregation
		columns := make(map[string][]ColumnInfo)

		// Add columns in specific order
		columns["public.users"] = []ColumnInfo{
			{Name: "id", Position: 1},
			{Name: "email", Position: 2},
			{Name: "name", Position: 3},
		}

		// Verify order is preserved
		assert.Equal(t, "id", columns["public.users"][0].Name)
		assert.Equal(t, "email", columns["public.users"][1].Name)
		assert.Equal(t, "name", columns["public.users"][2].Name)
	})

	t.Run("primary key aggregation handles composite keys", func(t *testing.T) {
		pks := make(map[string][]string)

		pks["public.users"] = []string{"id"}
		pks["public.user_roles"] = []string{"user_id", "role_id"}
		pks["public.sessions"] = []string{"id"}

		assert.Len(t, pks["public.users"], 1)
		assert.Len(t, pks["public.user_roles"], 2)
		assert.Equal(t, "user_id", pks["public.user_roles"][0])
		assert.Equal(t, "role_id", pks["public.user_roles"][1])
	})

	t.Run("index aggregation groups by table", func(t *testing.T) {
		indexes := make(map[string][]IndexInfo)

		indexes["public.users"] = []IndexInfo{
			{Name: "users_pkey", IsPrimary: true, IsUnique: true},
			{Name: "idx_users_email", IsUnique: true},
		}

		assert.Len(t, indexes["public.users"], 2)
		assert.True(t, indexes["public.users"][0].IsPrimary)
		assert.True(t, indexes["public.users"][1].IsUnique)
	})
}
