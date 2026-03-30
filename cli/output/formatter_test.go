package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  Format
		err   bool
	}{
		{"table", FormatTable, false},
		{"", FormatTable, false},
		{"TABLE", FormatTable, false},
		{"json", FormatJSON, false},
		{"JSON", FormatJSON, false},
		{"yaml", FormatYAML, false},
		{"YAML", FormatYAML, false},
		{"yml", FormatYAML, false},
		{"xml", "", true},
		{"csv", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func newTestFormatter(format Format) (*Formatter, *bytes.Buffer) {
	var buf bytes.Buffer
	f := NewFormatter(format, false, false)
	f.Writer = &buf
	return f, &buf
}

func TestFormatter_Print_JSON(t *testing.T) {
	f, buf := newTestFormatter(FormatJSON)

	data := map[string]string{"name": "test", "status": "active"}
	err := f.Print(data)
	require.NoError(t, err)

	var result map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, "active", result["status"])
}

func TestFormatter_Print_YAML(t *testing.T) {
	f, buf := newTestFormatter(FormatYAML)

	data := map[string]string{"name": "test", "status": "active"}
	err := f.Print(data)
	require.NoError(t, err)

	var result map[string]string
	require.NoError(t, yaml.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, "active", result["status"])
}

func TestFormatter_Print_Quiet(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatJSON, false, true)
	f.Writer = &buf

	err := f.Print(map[string]string{"key": "value"})
	assert.NoError(t, err)
	assert.Empty(t, buf.String())
}

func TestFormatter_Print_Slice(t *testing.T) {
	f, buf := newTestFormatter(FormatJSON)

	data := []string{"item1", "item2", "item3"}
	err := f.Print(data)
	require.NoError(t, err)

	var result []string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, []string{"item1", "item2", "item3"}, result)
}

func TestFormatter_PrintTable_Table(t *testing.T) {
	f, buf := newTestFormatter(FormatTable)

	f.PrintTable(TableData{
		Headers: []string{"Name", "Status"},
		Rows: [][]string{
			{"func1", "active"},
			{"func2", "inactive"},
		},
	})

	output := buf.String()
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "func1")
	assert.Contains(t, output, "func2")
}

func TestFormatter_PrintTable_NoHeaders(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatTable, true, false)
	f.Writer = &buf

	f.PrintTable(TableData{
		Headers: []string{"Name", "Status"},
		Rows: [][]string{
			{"func1", "active"},
		},
	})

	output := buf.String()
	assert.NotContains(t, output, "NAME")
	assert.Contains(t, output, "func1")
}

func TestFormatter_PrintTable_EmptyRows(t *testing.T) {
	f, buf := newTestFormatter(FormatTable)

	f.PrintTable(TableData{
		Headers: []string{"Name"},
		Rows:    [][]string{},
	})

	output := buf.String()
	assert.Contains(t, output, "NAME")
}

func TestFormatter_PrintTable_JSON(t *testing.T) {
	f, buf := newTestFormatter(FormatJSON)

	f.PrintTable(TableData{
		Headers: []string{"Name", "Status"},
		Rows: [][]string{
			{"func1", "active"},
		},
	})

	var result []map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 1)
	assert.Equal(t, "func1", result[0]["Name"])
	assert.Equal(t, "active", result[0]["Status"])
}

func TestFormatter_PrintTable_Quiet(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatTable, false, true)
	f.Writer = &buf

	f.PrintTable(TableData{
		Headers: []string{"Name"},
		Rows:    [][]string{{"func1"}},
	})

	assert.Empty(t, buf.String())
}

func TestFormatter_PrintSuccess(t *testing.T) {
	f, buf := newTestFormatter(FormatTable)

	f.PrintSuccess("Operation completed")
	assert.Contains(t, buf.String(), "Operation completed")
}

func TestFormatter_PrintSuccess_Quiet(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatTable, false, true)
	f.Writer = &buf

	f.PrintSuccess("Operation completed")
	assert.Empty(t, buf.String())
}

func TestFormatter_PrintInfo(t *testing.T) {
	f, buf := newTestFormatter(FormatTable)

	f.PrintInfo("Processing...")
	assert.Contains(t, buf.String(), "Processing...")
}

func TestFormatter_PrintInfo_Quiet(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatTable, false, true)
	f.Writer = &buf

	f.PrintInfo("Processing...")
	assert.Empty(t, buf.String())
}

func TestFormatter_PrintKeyValue(t *testing.T) {
	f, buf := newTestFormatter(FormatTable)

	f.PrintKeyValue("Name", "test-func")
	assert.Contains(t, buf.String(), "Name: test-func")
}

func TestFormatter_PrintKeyValue_JSON(t *testing.T) {
	f, buf := newTestFormatter(FormatJSON)

	f.PrintKeyValue("Name", "test-func")

	var result map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "test-func", result["Name"])
}

func TestFormatter_PrintKeyValue_YAML(t *testing.T) {
	f, buf := newTestFormatter(FormatYAML)

	f.PrintKeyValue("Name", "test-func")

	var result map[string]string
	require.NoError(t, yaml.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "test-func", result["Name"])
}

func TestFormatter_PrintKeyValue_Quiet(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatTable, false, true)
	f.Writer = &buf

	f.PrintKeyValue("Name", "test-func")
	assert.Empty(t, buf.String())
}

func TestFormatter_PrintList(t *testing.T) {
	f, buf := newTestFormatter(FormatTable)

	f.PrintList([]string{"item1", "item2", "item3"})
	output := buf.String()
	assert.Contains(t, output, "item1")
	assert.Contains(t, output, "item2")
	assert.Contains(t, output, "item3")
}

func TestFormatter_PrintList_JSON(t *testing.T) {
	f, buf := newTestFormatter(FormatJSON)

	f.PrintList([]string{"item1", "item2"})

	var result []string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, []string{"item1", "item2"}, result)
}

func TestFormatter_PrintList_Quiet(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(FormatTable, false, true)
	f.Writer = &buf

	f.PrintList([]string{"item1"})
	assert.Empty(t, buf.String())
}

func TestFormatter_PrintGeneric(t *testing.T) {
	// Default format (table) falls back to JSON for generic data
	f, buf := newTestFormatter(FormatTable)

	data := map[string]int{"count": 42}
	err := f.Print(data)
	require.NoError(t, err)

	var result map[string]int
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, 42, result["count"])
}
