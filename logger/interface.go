package logger

type Sqlable interface {
	InsertQuery() (query string, fields string, values []any)
	BufferSize() int
}

type PurgeQueryFunc func() string
type CreateTableQueryFunc func() string
type AlterTableQueryFunc func() string
