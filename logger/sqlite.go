package logger

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Sqlite struct {
	lock                     sync.Mutex
	db                       *sql.DB
	file                     string
	doInsert                 bool
	purgeQueryFuncList       []PurgeQueryFunc
	createTableQueryFuncList []CreateTableQueryFunc
}

func NewSqlite(file string, createTableQueryFuncList []CreateTableQueryFunc, purgeQueryFuncList []PurgeQueryFunc) *Sqlite {
	return &Sqlite{
		file:                     file,
		createTableQueryFuncList: createTableQueryFuncList,
		purgeQueryFuncList:       purgeQueryFuncList,
	}
}

func (s *Sqlite) Init(logTTL time.Duration) error {
	fmt.Println("initializing database:", s.file)
	db, err := sql.Open("sqlite", s.file)
	if err != nil {
		return fmt.Errorf("opening database: %s", err.Error())
	}

	for _, createQuery := range s.createTableQueryFuncList {
		if _, err := db.Exec(createQuery()); err != nil {
			return fmt.Errorf("creating table: %s", err.Error())
		}
	}

	fmt.Println("database initialized, will purge every:", logTTL.String())

	if logTTL > 0 {
		go func() {
			for {
				time.Sleep(time.Minute)
				err := s.Purge(logTTL)
				if err != nil {
					panic(fmt.Errorf("purging database: %s", err.Error()))
				}
			}
		}()
	}

	s.db = db

	return nil
}

func (s *Sqlite) Purge(ttl time.Duration) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	t := time.Now().Add(ttl * -1)
	fmt.Println("purging database older than:", t)
	for _, purgeQueryFunc := range s.purgeQueryFuncList {
		if res, err := s.db.Exec(purgeQueryFunc(), t); err != nil {
			return err
		} else {
			c, _ := res.RowsAffected()
			fmt.Println("purged rows:", c)
		}
	}

	return nil
}

func (s *Sqlite) Log(data Sqlable) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	insertQuery, params := data.InsertQuery()

	_, err := s.db.Exec(insertQuery, params...)
	if err != nil {
		return fmt.Errorf("inserting data: %s", err.Error())
	}
	return nil
}

func (s *Sqlite) SingleRowQuery(sql string, handleRow func(row *sql.Rows) error, params ...any) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	rows, err := s.db.Query(sql, params...)
	if err != nil {
		return fmt.Errorf("querying last position: %s", err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := handleRow(rows)
		if err != nil {
			return fmt.Errorf("handling row: %s", err.Error())
		}
		return nil
	}

	return nil
}

func (s *Sqlite) Query(sql string, handleRow func(row *sql.Rows) error, params []any) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	fmt.Println("Running query:", sql, params)

	rows, err := s.db.Query(sql, params...)
	if err != nil {
		return fmt.Errorf("querying last position: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := handleRow(rows)
		if err != nil {
			return fmt.Errorf("handling row: %s", err.Error())
		}
	}

	return nil
}
