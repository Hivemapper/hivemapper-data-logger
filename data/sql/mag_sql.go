package sql

import (
	"time"

	"github.com/Hivemapper/hivemapper-data-logger/data/session"
)

const MagCreateTable string = `
  CREATE TABLE IF NOT EXISTS magnetometer (
  	id INTEGER NOT NULL PRIMARY KEY,
  	system_time TIMESTAMP NOT NULL,
	mag_x REAL NOT NULL,
	mag_y REAL NOT NULL,
	mag_z REAL NOT NULL
  );
	create index if not exists mag_time_idx on magnetometer(system_time);
`

const MagAlterTable string = `
	ALTER TABLE magnetometer ADD COLUMN session TEXT NOT NULL DEFAULT '';
`

const insertMagnetometerQuery string = `INSERT INTO magnetometer VALUES `
const insertMagnetometerFields string = `(NULL,?,?,?,?,?),`

const purgeQuery string = `
	DELETE FROM magnetometer WHERE rowid NOT IN (
		SELECT rowid FROM gnss ORDER BY rowid DESC LIMIT 500000
	);
`

type MagnetometerSqlWrapper struct {
	System_time time.Time
	Mag_x       float64
	Mag_y       float64
	Mag_z       float64
}

func NewMagnetometerSqlWrapper(system_time time.Time, mag_x float64, mag_y float64, mag_z float64) *MagnetometerSqlWrapper {
	return &MagnetometerSqlWrapper{
		System_time: system_time,
		Mag_x:       mag_x,
		Mag_y:       mag_y,
		Mag_z:       mag_z,
	}
}

func (s *MagnetometerSqlWrapper) InsertQuery() (string, string, []any) {
	sessionID, err := session.GetSession()
	if err != nil {
		panic(err) // Handle error if any
	}
	return insertMagnetometerQuery, insertMagnetometerFields, []any{
		s.System_time.Format("2006-01-02 15:04:05.99999"),
		s.Mag_x,
		s.Mag_y,
		s.Mag_z,
		sessionID,
	}
}

func MagCreateTableQuery() string {
	return MagCreateTable
}

func MagAlterTableQuery() string {
	return MagAlterTable
}

func MagPurgeQuery() string {
	return purgeQuery
}
