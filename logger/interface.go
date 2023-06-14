package logger

import "database/sql"

type Sqlable interface {
	InsertQuery() (*sql.Stmt, []any)
}

type PurgeQueryFunc func() string
type CreateTableQueryFunc func() string
