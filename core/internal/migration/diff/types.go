package diff

type Schema struct {
	Tables      map[string]Table
	Indexes     map[string]Index
	Constraints map[string]Constraint
	Enums       map[string]Enum
	ForeignKeys map[string]ForeignKey
}

type Table struct {
	Name    string
	Columns map[string]Column
}

type Column struct {
	Name       string
	DataType   string
	Nullable   bool
	DefaultSQL string
}

type Index struct {
	Name      string
	TableName string
	Def       string
}

type Constraint struct {
	Name      string
	TableName string
	Def       string
}

type Enum struct {
	Name   string
	Values []string
}

type ForeignKey struct {
	Name      string
	TableName string
	Def       string
}

type Change struct {
	Type        string
	Description string
	SQL         string
	ReverseSQL  string
}
