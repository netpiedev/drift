package diff

import (
	"context"
	"database/sql"
	"fmt"
)

func Introspect(ctx context.Context, db *sql.DB, schemaName string) (Schema, error) {
	if schemaName == "" {
		schemaName = "public"
	}

	s := Schema{
		Tables:      map[string]Table{},
		Indexes:     map[string]Index{},
		Constraints: map[string]Constraint{},
		Enums:       map[string]Enum{},
		ForeignKeys: map[string]ForeignKey{},
	}

	if err := introspectTables(ctx, db, schemaName, &s); err != nil {
		return Schema{}, err
	}
	if err := introspectIndexes(ctx, db, schemaName, &s); err != nil {
		return Schema{}, err
	}
	if err := introspectConstraints(ctx, db, schemaName, &s); err != nil {
		return Schema{}, err
	}
	if err := introspectEnums(ctx, db, schemaName, &s); err != nil {
		return Schema{}, err
	}
	if err := introspectForeignKeys(ctx, db, schemaName, &s); err != nil {
		return Schema{}, err
	}
	return s, nil
}

func introspectTables(ctx context.Context, db *sql.DB, schemaName string, s *Schema) error {
	const q = `
SELECT table_name, column_name, data_type, is_nullable, COALESCE(column_default,'')
FROM information_schema.columns
WHERE table_schema = $1
ORDER BY table_name, ordinal_position;`
	rows, err := db.QueryContext(ctx, q, schemaName)
	if err != nil {
		return fmt.Errorf("introspect tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, colName, dataType, isNullable, defaultSQL string
		if err := rows.Scan(&tableName, &colName, &dataType, &isNullable, &defaultSQL); err != nil {
			return fmt.Errorf("scan column metadata: %w", err)
		}
		tbl, ok := s.Tables[tableName]
		if !ok {
			tbl = Table{Name: tableName, Columns: map[string]Column{}}
		}
		tbl.Columns[colName] = Column{
			Name:       colName,
			DataType:   dataType,
			Nullable:   isNullable == "YES",
			DefaultSQL: defaultSQL,
		}
		s.Tables[tableName] = tbl
	}
	return rows.Err()
}

func introspectIndexes(ctx context.Context, db *sql.DB, schemaName string, s *Schema) error {
	const q = `
SELECT indexname, tablename, indexdef
FROM pg_indexes
WHERE schemaname = $1;`
	rows, err := db.QueryContext(ctx, q, schemaName)
	if err != nil {
		return fmt.Errorf("introspect indexes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var idx Index
		if err := rows.Scan(&idx.Name, &idx.TableName, &idx.Def); err != nil {
			return fmt.Errorf("scan index: %w", err)
		}
		s.Indexes[idx.Name] = idx
	}
	return rows.Err()
}

func introspectConstraints(ctx context.Context, db *sql.DB, schemaName string, s *Schema) error {
	const q = `
SELECT c.conname,
       t.relname AS table_name,
       pg_get_constraintdef(c.oid)
FROM pg_constraint c
JOIN pg_class t ON t.oid = c.conrelid
JOIN pg_namespace n ON n.oid = t.relnamespace
WHERE n.nspname = $1
  AND c.contype IN ('c','u','p');`
	rows, err := db.QueryContext(ctx, q, schemaName)
	if err != nil {
		return fmt.Errorf("introspect constraints: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c Constraint
		if err := rows.Scan(&c.Name, &c.TableName, &c.Def); err != nil {
			return fmt.Errorf("scan constraint: %w", err)
		}
		s.Constraints[c.Name] = c
	}
	return rows.Err()
}

func introspectEnums(ctx context.Context, db *sql.DB, schemaName string, s *Schema) error {
	const q = `
SELECT t.typname, e.enumlabel
FROM pg_type t
JOIN pg_enum e ON e.enumtypid = t.oid
JOIN pg_namespace n ON n.oid = t.typnamespace
WHERE n.nspname = $1
ORDER BY t.typname, e.enumsortorder;`
	rows, err := db.QueryContext(ctx, q, schemaName)
	if err != nil {
		return fmt.Errorf("introspect enums: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var enumName string
		var label string
		if err := rows.Scan(&enumName, &label); err != nil {
			return fmt.Errorf("scan enum: %w", err)
		}
		e, ok := s.Enums[enumName]
		if !ok {
			e = Enum{Name: enumName}
		}
		e.Values = append(e.Values, label)
		s.Enums[enumName] = e
	}
	return rows.Err()
}

func introspectForeignKeys(ctx context.Context, db *sql.DB, schemaName string, s *Schema) error {
	const q = `
SELECT c.conname,
       t.relname AS table_name,
       pg_get_constraintdef(c.oid)
FROM pg_constraint c
JOIN pg_class t ON t.oid = c.conrelid
JOIN pg_namespace n ON n.oid = t.relnamespace
WHERE n.nspname = $1
  AND c.contype = 'f';`
	rows, err := db.QueryContext(ctx, q, schemaName)
	if err != nil {
		return fmt.Errorf("introspect foreign keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var fk ForeignKey
		if err := rows.Scan(&fk.Name, &fk.TableName, &fk.Def); err != nil {
			return fmt.Errorf("scan foreign key: %w", err)
		}
		s.ForeignKeys[fk.Name] = fk
	}
	return rows.Err()
}
