package logger

type Sqlable interface {
	InsertQuery() (string, []any)
}

type PurgeQueryFunc func() string
type CreateTableQueryFunc func() string
