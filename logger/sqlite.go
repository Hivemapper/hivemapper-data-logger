package logger

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Sqlite struct {
	lock                 sync.Mutex
	db                   *sql.DB
	file                 string
	doInsert             bool
	purgeQueryFunc       PurgeQueryFunc
	createTableQueryFunc CreateTableQueryFunc
}

func NewSqlite(file string, createTableQueryFunc CreateTableQueryFunc, purgeQueryFunc PurgeQueryFunc) *Sqlite {
	return &Sqlite{
		file:                 file,
		createTableQueryFunc: createTableQueryFunc,
		purgeQueryFunc:       purgeQueryFunc,
	}
}

func (s *Sqlite) Init(logTTL time.Duration) error {
	fmt.Println("initializing database:", s.file)
	db, err := sql.Open("sqlite", s.file)
	if err != nil {
		return fmt.Errorf("opening database: %s", err.Error())
	}

	if _, err := db.Exec(s.createTableQueryFunc()); err != nil {
		return fmt.Errorf("creating table: %s", err.Error())
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
	if res, err := s.db.Exec(s.purgeQueryFunc(), t); err != nil {
		return err
	} else {
		c, _ := res.RowsAffected()
		fmt.Println("purged rows:", c)
	}

	return nil
}
func (s *Sqlite) StartStoring() {
	s.doInsert = true
}

func (s *Sqlite) Log(data Sqlable) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.doInsert {
		return nil
	}

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

	fmt.Println("Running query:", sql, params)

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
