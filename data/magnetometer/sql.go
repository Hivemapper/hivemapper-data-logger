package magnetometer

import (
	"time"
)

const MagCreateTable string = `
  CREATE TABLE IF NOT EXISTS magnetometer (
  	id INTEGER NOT NULL PRIMARY KEY,
  	system_time TIMESTAMP NOT NULL,
	mag_x TEXT NOT NULL,
	mag_y INTEGER NOT NULL,
	mag_z INTEGER NOT NULL
  );
	create index if not exists mag_time_idx on magnetometer(system_time);
`

const insertMagnetometerQuery string = `INSERT INTO magnetometer VALUES `
const insertMagnetometerFields string = `(NULL,?,?,?,?),`

const purgeQuery string = `
	DELETE FROM magnetometer WHERE system_time < ?;
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
	return insertMagnetometerQuery, insertMagnetometerFields, []any{
		s.System_time.Format("2006-01-02 15:04:05.99999"),
		s.Mag_x,
		s.Mag_y,
		s.Mag_z,
	}
}

func CreateTableQuery() string {
	return MagCreateTable
}

func PurgeQuery() string {
	return purgeQuery
}
