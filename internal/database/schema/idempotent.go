package schema

import (
	"fmt"
	"strings"

	"github.com/pgplex/pgparser/nodes"
	"github.com/pgplex/pgparser/parser"
	"github.com/rs/zerolog/log"
)

// MakeSQLIdempotent transforms SQL to be idempotent by prepending DROP IF EXISTS
// statements before CREATE POLICY and ALTER TABLE ADD CONSTRAINT statements.
// This allows schema SQL files to be safely re-applied to existing databases.
func MakeSQLIdempotent(sql string) string {
	stmts, err := parser.Parse(sql)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse SQL for idempotency transformation, using original")
		return sql
	}

	if stmts == nil || len(stmts.Items) == 0 {
		return sql
	}

	type dropInfo struct {
		pattern  string
		dropSQL  string
		foundPos int
	}
	var drops []dropInfo

	for _, item := range stmts.Items {
		switch stmt := item.(type) {
		case *nodes.CreatePolicyStmt:
			if stmt.Table != nil {
				tableName := formatRangeVar(stmt.Table)
				dropSQL := fmt.Sprintf("DROP POLICY IF EXISTS %s ON %s CASCADE;\n", quoteIdent(stmt.PolicyName), tableName)
				patternQuoted := "CREATE POLICY \"" + stmt.PolicyName + "\""
				patternUnquoted := "CREATE POLICY " + stmt.PolicyName
				drops = append(drops, dropInfo{pattern: patternQuoted, dropSQL: dropSQL, foundPos: -1})
				drops = append(drops, dropInfo{pattern: patternUnquoted, dropSQL: dropSQL, foundPos: -1})
			}

		case *nodes.AlterTableStmt:
			if stmt.Cmds != nil && stmt.Relation != nil {
				for _, cmd := range stmt.Cmds.Items {
					alterCmd, ok := cmd.(*nodes.AlterTableCmd)
					if !ok {
						continue
					}
					if alterCmd.Subtype == 17 && alterCmd.Def != nil {
						if constraint, ok := alterCmd.Def.(*nodes.Constraint); ok && constraint.Conname != "" {
							tableName := formatRangeVar(stmt.Relation)
							dropSQL := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s;\n", tableName, quoteIdent(constraint.Conname))
							pattern := "ALTER TABLE " + stmt.Relation.Relname
							drops = append(drops, dropInfo{pattern: pattern, dropSQL: dropSQL, foundPos: -1})
							break
						}
					}
				}
			}
		}
	}

	if len(drops) == 0 {
		return sql
	}

	upperSQL := strings.ToUpper(sql)
	lastFoundPos := make(map[string]int)

	for i := range drops {
		upperPattern := strings.ToUpper(drops[i].pattern)
		searchStart := lastFoundPos[upperPattern]
		idx := strings.Index(upperSQL[searchStart:], upperPattern)
		if idx != -1 {
			drops[i].foundPos = searchStart + idx
			lastFoundPos[upperPattern] = drops[i].foundPos + len(upperPattern)
		}
	}

	for i := 0; i < len(drops)-1; i++ {
		for j := i + 1; j < len(drops); j++ {
			if drops[i].foundPos < drops[j].foundPos {
				drops[i], drops[j] = drops[j], drops[i]
			}
		}
	}

	result := sql
	for _, drop := range drops {
		if drop.foundPos >= 0 {
			result = result[:drop.foundPos] + drop.dropSQL + result[drop.foundPos:]
		}
	}

	return result
}

func formatRangeVar(rv *nodes.RangeVar) string {
	if rv.Schemaname != "" {
		return fmt.Sprintf("%s.%s", quoteIdent(rv.Schemaname), quoteIdent(rv.Relname))
	}
	return quoteIdent(rv.Relname)
}

func quoteIdent(name string) string {
	if strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`) {
		return name
	}
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}
