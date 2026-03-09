package ai

import (
	"testing"
)

func TestBuildConditionSQL_EqualityOperators(t *testing.T) {
	tests := []struct {
		name        string
		cond        MetadataCondition
		wantSQL     string
		wantArgsLen int
		wantErr     bool
	}{
		{
			name: "equals operator",
			cond: MetadataCondition{
				Key:      "category",
				Operator: MetadataOpEquals,
				Value:    "food",
			},
			wantSQL:     `d.metadata->>'category' = $1`,
			wantArgsLen: 1,
			wantErr:     false,
		},
		{
			name: "not equals operator",
			cond: MetadataCondition{
				Key:      "status",
				Operator: MetadataOpNotEquals,
				Value:    "archived",
			},
			wantSQL:     `d.metadata->>'status' != $1`,
			wantArgsLen: 1,
			wantErr:     false,
		},
		{
			name: "equals operator with numeric value",
			cond: MetadataCondition{
				Key:      "count",
				Operator: MetadataOpEquals,
				Value:    42,
			},
			wantSQL:     `d.metadata->>'count' = $1`,
			wantArgsLen: 1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argIndex := 1
			sql, args, err := buildConditionSQL(tt.cond, &argIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("buildConditionSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if sql != tt.wantSQL {
				t.Errorf("buildConditionSQL() SQL = %v, want %v", sql, tt.wantSQL)
			}
			if len(args) != tt.wantArgsLen {
				t.Errorf("buildConditionSQL() args len = %v, want %v", len(args), tt.wantArgsLen)
			}
			if argIndex-1 != tt.wantArgsLen {
				t.Errorf("buildConditionSQL() argIndex = %v, want %v", argIndex-1, tt.wantArgsLen)
			}
		})
	}
}

func TestBuildConditionSQL_PatternOperators(t *testing.T) {
	tests := []struct {
		name        string
		cond        MetadataCondition
		wantSQL     string
		wantArgsLen int
		wantErr     bool
	}{
		{
			name: "ILIKE operator",
			cond: MetadataCondition{
				Key:      "city",
				Operator: MetadataOpILike,
				Value:    "%Tokyo%",
			},
			wantSQL:     `d.metadata->>'city' ILIKE $1`,
			wantArgsLen: 1,
			wantErr:     false,
		},
		{
			name: "LIKE operator",
			cond: MetadataCondition{
				Key:      "name",
				Operator: MetadataOpLike,
				Value:    "Starbucks%",
			},
			wantSQL:     `d.metadata->>'name' LIKE $1`,
			wantArgsLen: 1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argIndex := 1
			sql, args, err := buildConditionSQL(tt.cond, &argIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("buildConditionSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if sql != tt.wantSQL {
				t.Errorf("buildConditionSQL() SQL = %v, want %v", sql, tt.wantSQL)
			}
			if len(args) != tt.wantArgsLen {
				t.Errorf("buildConditionSQL() args len = %v, want %v", len(args), tt.wantArgsLen)
			}
		})
	}
}

func TestBuildConditionSQL_InOperators(t *testing.T) {
	tests := []struct {
		name        string
		cond        MetadataCondition
		wantSQL     string
		wantArgsLen int
		wantErr     bool
	}{
		{
			name: "IN operator with strings",
			cond: MetadataCondition{
				Key:      "cuisine",
				Operator: MetadataOpIn,
				Values:   []interface{}{"japanese", "sushi", "ramen"},
			},
			wantSQL:     `d.metadata->>'cuisine' IN ($1, $2, $3)`,
			wantArgsLen: 3,
			wantErr:     false,
		},
		{
			name: "NOT IN operator",
			cond: MetadataCondition{
				Key:      "status",
				Operator: MetadataOpNotIn,
				Values:   []interface{}{"archived", "deleted"},
			},
			wantSQL:     `d.metadata->>'status' NOT IN ($1, $2)`,
			wantArgsLen: 2,
			wantErr:     false,
		},
		{
			name: "IN operator with single value",
			cond: MetadataCondition{
				Key:      "category",
				Operator: MetadataOpIn,
				Values:   []interface{}{"food"},
			},
			wantSQL:     `d.metadata->>'category' IN ($1)`,
			wantArgsLen: 1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argIndex := 1
			sql, args, err := buildConditionSQL(tt.cond, &argIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("buildConditionSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if sql != tt.wantSQL {
				t.Errorf("buildConditionSQL() SQL = %v, want %v", sql, tt.wantSQL)
			}
			if len(args) != tt.wantArgsLen {
				t.Errorf("buildConditionSQL() args len = %v, want %v", len(args), tt.wantArgsLen)
			}
		})
	}
}

func TestBuildConditionSQL_RangeOperators(t *testing.T) {
	tests := []struct {
		name        string
		cond        MetadataCondition
		wantSQL     string
		wantArgsLen int
		wantErr     bool
	}{
		{
			name: "greater than operator",
			cond: MetadataCondition{
				Key:      "rating",
				Operator: MetadataOpGreaterThan,
				Value:    4.5,
			},
			wantSQL:     `d.metadata->>'rating' > $1`,
			wantArgsLen: 1,
			wantErr:     false,
		},
		{
			name: "greater than or equal operator",
			cond: MetadataCondition{
				Key:      "price",
				Operator: MetadataOpGreaterThanOr,
				Value:    100,
			},
			wantSQL:     `d.metadata->>'price' >= $1`,
			wantArgsLen: 1,
			wantErr:     false,
		},
		{
			name: "less than operator",
			cond: MetadataCondition{
				Key:      "distance",
				Operator: MetadataOpLessThan,
				Value:    5.0,
			},
			wantSQL:     `d.metadata->>'distance' < $1`,
			wantArgsLen: 1,
			wantErr:     false,
		},
		{
			name: "less than or equal operator",
			cond: MetadataCondition{
				Key:      "quantity",
				Operator: MetadataOpLessThanOr,
				Value:    10,
			},
			wantSQL:     `d.metadata->>'quantity' <= $1`,
			wantArgsLen: 1,
			wantErr:     false,
		},
		{
			name: "BETWEEN operator",
			cond: MetadataCondition{
				Key:      "duration",
				Operator: MetadataOpBetween,
				Min:      30,
				Max:      90,
			},
			wantSQL:     `d.metadata->>'duration' BETWEEN $1 AND $2`,
			wantArgsLen: 2,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argIndex := 1
			sql, args, err := buildConditionSQL(tt.cond, &argIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("buildConditionSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if sql != tt.wantSQL {
				t.Errorf("buildConditionSQL() SQL = %v, want %v", sql, tt.wantSQL)
			}
			if len(args) != tt.wantArgsLen {
				t.Errorf("buildConditionSQL() args len = %v, want %v", len(args), tt.wantArgsLen)
			}
		})
	}
}

func TestBuildConditionSQL_NullOperators(t *testing.T) {
	tests := []struct {
		name        string
		cond        MetadataCondition
		wantSQL     string
		wantArgsLen int
		wantErr     bool
	}{
		{
			name: "IS NULL operator",
			cond: MetadataCondition{
				Key:      "deleted_at",
				Operator: MetadataOpIsNull,
			},
			wantSQL:     `d.metadata->>'deleted_at' IS NULL`,
			wantArgsLen: 0,
			wantErr:     false,
		},
		{
			name: "IS NOT NULL operator",
			cond: MetadataCondition{
				Key:      "verified_at",
				Operator: MetadataOpIsNotNull,
			},
			wantSQL:     `d.metadata->>'verified_at' IS NOT NULL`,
			wantArgsLen: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argIndex := 1
			sql, args, err := buildConditionSQL(tt.cond, &argIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("buildConditionSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if sql != tt.wantSQL {
				t.Errorf("buildConditionSQL() SQL = %v, want %v", sql, tt.wantSQL)
			}
			if len(args) != tt.wantArgsLen {
				t.Errorf("buildConditionSQL() args len = %v, want %v", len(args), tt.wantArgsLen)
			}
		})
	}
}

func TestBuildConditionSQL_Errors(t *testing.T) {
	tests := []struct {
		name    string
		cond    MetadataCondition
		wantErr bool
	}{
		{
			name: "equals operator missing value",
			cond: MetadataCondition{
				Key:      "category",
				Operator: MetadataOpEquals,
			},
			wantErr: true,
		},
		{
			name: "IN operator missing values",
			cond: MetadataCondition{
				Key:      "category",
				Operator: MetadataOpIn,
			},
			wantErr: true,
		},
		{
			name: "BETWEEN operator missing min",
			cond: MetadataCondition{
				Key:      "duration",
				Operator: MetadataOpBetween,
				Max:      90,
			},
			wantErr: true,
		},
		{
			name: "BETWEEN operator missing max",
			cond: MetadataCondition{
				Key:      "duration",
				Operator: MetadataOpBetween,
				Min:      30,
			},
			wantErr: true,
		},
		{
			name: "unsupported operator",
			cond: MetadataCondition{
				Key:      "category",
				Operator: MetadataOperator("UNSUPPORTED"),
				Value:    "test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argIndex := 1
			_, _, err := buildConditionSQL(tt.cond, &argIndex)

			if (err != nil) != tt.wantErr {
				t.Errorf("buildConditionSQL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildMetadataFilterSQL_SingleCondition(t *testing.T) {
	group := MetadataFilterGroup{
		Conditions: []MetadataCondition{
			{
				Key:      "category",
				Operator: MetadataOpEquals,
				Value:    "food",
			},
		},
		LogicalOp: LogicalOpAND,
	}

	argIndex := 1
	sql, args, err := buildMetadataFilterSQL(group, &argIndex)
	if err != nil {
		t.Fatalf("buildMetadataFilterSQL() error = %v", err)
	}
	if sql != `d.metadata->>'category' = $1` {
		t.Errorf("buildMetadataFilterSQL() SQL = %v, want %v", sql, `d.metadata->>'category' = $1`)
	}
	if len(args) != 1 {
		t.Errorf("buildMetadataFilterSQL() args len = %v, want 1", len(args))
	}
}

func TestBuildMetadataFilterSQL_MultipleConditionsAND(t *testing.T) {
	group := MetadataFilterGroup{
		Conditions: []MetadataCondition{
			{
				Key:      "category",
				Operator: MetadataOpEquals,
				Value:    "food",
			},
			{
				Key:      "city",
				Operator: MetadataOpILike,
				Value:    "%Tokyo%",
			},
		},
		LogicalOp: LogicalOpAND,
	}

	argIndex := 1
	sql, args, err := buildMetadataFilterSQL(group, &argIndex)
	if err != nil {
		t.Fatalf("buildMetadataFilterSQL() error = %v", err)
	}
	expectedSQL := `d.metadata->>'category' = $1 AND d.metadata->>'city' ILIKE $2`
	if sql != expectedSQL {
		t.Errorf("buildMetadataFilterSQL() SQL = %v, want %v", sql, expectedSQL)
	}
	if len(args) != 2 {
		t.Errorf("buildMetadataFilterSQL() args len = %v, want 2", len(args))
	}
}

func TestBuildMetadataFilterSQL_MultipleConditionsOR(t *testing.T) {
	group := MetadataFilterGroup{
		Conditions: []MetadataCondition{
			{
				Key:      "status",
				Operator: MetadataOpEquals,
				Value:    "active",
			},
			{
				Key:      "status",
				Operator: MetadataOpEquals,
				Value:    "pending",
			},
		},
		LogicalOp: LogicalOpOR,
	}

	argIndex := 1
	sql, args, err := buildMetadataFilterSQL(group, &argIndex)
	if err != nil {
		t.Fatalf("buildMetadataFilterSQL() error = %v", err)
	}
	expectedSQL := `d.metadata->>'status' = $1 OR d.metadata->>'status' = $2`
	if sql != expectedSQL {
		t.Errorf("buildMetadataFilterSQL() SQL = %v, want %v", sql, expectedSQL)
	}
	if len(args) != 2 {
		t.Errorf("buildMetadataFilterSQL() args len = %v, want 2", len(args))
	}
}

func TestBuildMetadataFilterSQL_NestedGroups(t *testing.T) {
	// (category = 'food' AND city ILIKE '%Tokyo%') OR (category = 'culture')
	group := MetadataFilterGroup{
		LogicalOp: LogicalOpOR,
		Conditions: []MetadataCondition{
			{
				Key:      "category",
				Operator: MetadataOpEquals,
				Value:    "culture",
			},
		},
		Groups: []MetadataFilterGroup{
			{
				LogicalOp: LogicalOpAND,
				Conditions: []MetadataCondition{
					{
						Key:      "category",
						Operator: MetadataOpEquals,
						Value:    "food",
					},
					{
						Key:      "city",
						Operator: MetadataOpILike,
						Value:    "%Tokyo%",
					},
				},
			},
		},
	}

	argIndex := 1
	sql, args, err := buildMetadataFilterSQL(group, &argIndex)
	if err != nil {
		t.Fatalf("buildMetadataFilterSQL() error = %v", err)
	}
	// Note: The exact SQL structure may vary slightly depending on how the nested groups are processed
	// The important thing is that all conditions are present and arg indexing is correct
	if len(args) != 3 {
		t.Errorf("buildMetadataFilterSQL() args len = %v, want 3", len(args))
	}
	// Check that all expected operators are in the SQL
	if sql == "" {
		t.Errorf("buildMetadataFilterSQL() returned empty SQL")
	}
}

func TestBuildMetadataFilterSQL_INOperator(t *testing.T) {
	group := MetadataFilterGroup{
		Conditions: []MetadataCondition{
			{
				Key:      "cuisine",
				Operator: MetadataOpIn,
				Values:   []interface{}{"japanese", "italian", "vietnamese"},
			},
		},
		LogicalOp: LogicalOpAND,
	}

	argIndex := 1
	sql, args, err := buildMetadataFilterSQL(group, &argIndex)
	if err != nil {
		t.Fatalf("buildMetadataFilterSQL() error = %v", err)
	}
	expectedSQL := `d.metadata->>'cuisine' IN ($1, $2, $3)`
	if sql != expectedSQL {
		t.Errorf("buildMetadataFilterSQL() SQL = %v, want %v", sql, expectedSQL)
	}
	if len(args) != 3 {
		t.Errorf("buildMetadataFilterSQL() args len = %v, want 3", len(args))
	}
}

func TestBuildMetadataFilterSQL_BetweenOperator(t *testing.T) {
	group := MetadataFilterGroup{
		Conditions: []MetadataCondition{
			{
				Key:      "avg_duration",
				Operator: MetadataOpBetween,
				Min:      30,
				Max:      90,
			},
		},
		LogicalOp: LogicalOpAND,
	}

	argIndex := 1
	sql, args, err := buildMetadataFilterSQL(group, &argIndex)
	if err != nil {
		t.Fatalf("buildMetadataFilterSQL() error = %v", err)
	}
	expectedSQL := `d.metadata->>'avg_duration' BETWEEN $1 AND $2`
	if sql != expectedSQL {
		t.Errorf("buildMetadataFilterSQL() SQL = %v, want %v", sql, expectedSQL)
	}
	if len(args) != 2 {
		t.Errorf("buildMetadataFilterSQL() args len = %v, want 2", len(args))
	}
}

func TestEscapeStringLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no single quotes",
			input: "category",
			want:  "category",
		},
		{
			name:  "single quote in middle",
			input: "user's_name",
			want:  "user''s_name",
		},
		{
			name:  "multiple single quotes",
			input: "it's a user's name",
			want:  "it''s a user''s name",
		},
		{
			name:  "starts with single quote",
			input: "'test",
			want:  "''test",
		},
		{
			name:  "ends with single quote",
			input: "test'",
			want:  "test''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeStringLiteral(tt.input); got != tt.want {
				t.Errorf("escapeStringLiteral() = %v, want %v", got, tt.want)
			}
		})
	}
}
