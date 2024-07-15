package logger

type Sqlable interface {
	InsertQuery() (query string, fields string, values []any)
}

type CreateTableQueryFunc func() string
type AlterTableQueryFunc func() string
