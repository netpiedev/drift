package diff

import "fmt"

func Compare(from Schema, to Schema) []Change {
	changes := make([]Change, 0)

	for tableName, toTable := range to.Tables {
		fromTable, exists := from.Tables[tableName]
		if !exists {
			changes = append(changes, Change{
				Type:        "new_table",
				Description: fmt.Sprintf("table %s exists in target but not source", tableName),
				SQL:         generateCreateTable(toTable),
				ReverseSQL:  fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName),
			})
			continue
		}

		for colName, toCol := range toTable.Columns {
			fromCol, colExists := fromTable.Columns[colName]
			if !colExists {
				changes = append(changes, Change{
					Type:        "new_column",
					Description: fmt.Sprintf("column %s.%s exists in target but not source", tableName, colName),
					SQL:         fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", tableName, colName, toCol.DataType),
					ReverseSQL:  fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s;", tableName, colName),
				})
				continue
			}
			if fromCol.DataType != toCol.DataType || fromCol.Nullable != toCol.Nullable || fromCol.DefaultSQL != toCol.DefaultSQL {
				changes = append(changes, Change{
					Type:        "column_change",
					Description: fmt.Sprintf("column changed: %s.%s", tableName, colName),
					SQL:         fmt.Sprintf("-- manual review required for column change %s.%s", tableName, colName),
					ReverseSQL:  fmt.Sprintf("-- manual rollback required for column change %s.%s", tableName, colName),
				})
			}
		}

		for colName := range fromTable.Columns {
			if _, ok := toTable.Columns[colName]; !ok {
				changes = append(changes, Change{
					Type:        "removed_column",
					Description: fmt.Sprintf("column %s.%s exists in source but not target", tableName, colName),
					SQL:         fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", tableName, colName),
					ReverseSQL:  fmt.Sprintf("-- manual rollback required to recreate %s.%s", tableName, colName),
				})
			}
		}
	}

	for tableName := range from.Tables {
		if _, ok := to.Tables[tableName]; !ok {
			changes = append(changes, Change{
				Type:        "removed_table",
				Description: fmt.Sprintf("table %s exists in source but not target", tableName),
				SQL:         fmt.Sprintf("DROP TABLE %s;", tableName),
				ReverseSQL:  fmt.Sprintf("-- manual rollback required to recreate table %s", tableName),
			})
		}
	}

	changes = append(changes, compareSimpleMap("index", from.Indexes, to.Indexes)...)
	changes = append(changes, compareSimpleMap("constraint", from.Constraints, to.Constraints)...)
	changes = append(changes, compareSimpleMap("enum", from.Enums, to.Enums)...)
	changes = append(changes, compareSimpleMap("foreign_key", from.ForeignKeys, to.ForeignKeys)...)

	return changes
}

func compareSimpleMap[T any](kind string, from map[string]T, to map[string]T) []Change {
	changes := make([]Change, 0)
	for name := range to {
		if _, ok := from[name]; !ok {
			changes = append(changes, Change{
				Type:        "new_" + kind,
				Description: fmt.Sprintf("%s %s exists in target but not source", kind, name),
				SQL:         fmt.Sprintf("-- create %s %s", kind, name),
				ReverseSQL:  fmt.Sprintf("-- drop %s %s", kind, name),
			})
		}
	}
	for name := range from {
		if _, ok := to[name]; !ok {
			changes = append(changes, Change{
				Type:        "removed_" + kind,
				Description: fmt.Sprintf("%s %s exists in source but not target", kind, name),
				SQL:         fmt.Sprintf("-- drop %s %s", kind, name),
				ReverseSQL:  fmt.Sprintf("-- recreate %s %s", kind, name),
			})
		}
	}
	return changes
}

func generateCreateTable(t Table) string {
	if len(t.Columns) == 0 {
		return fmt.Sprintf("CREATE TABLE %s ();", t.Name)
	}
	first := true
	sql := fmt.Sprintf("CREATE TABLE %s (", t.Name)
	for _, c := range t.Columns {
		if !first {
			sql += ", "
		}
		first = false
		nullable := ""
		if !c.Nullable {
			nullable = " NOT NULL"
		}
		sql += fmt.Sprintf("%s %s%s", c.Name, c.DataType, nullable)
	}
	sql += ");"
	return sql
}
