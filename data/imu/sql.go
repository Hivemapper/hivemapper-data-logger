package imu

import (
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
)

const MergedCreateTable string = `
  CREATE TABLE IF NOT EXISTS direction_events (
  	id INTEGER NOT NULL PRIMARY KEY,
	time DATETIME NOT NULL,
	name TEXT NOT NULL,
	gnss_latitude REAL NOT NULL,
	gnss_longitude REAL NOT NULL,
	gnss_speed REAL NOT NULL
  );`

const insertQuery string = `
INSERT INTO direction_events VALUES(NULL,?,?,?,?,?);
`

const purgeQuery string = `
DELETE FROM direction_events WHERE time < ?;
`

func CreateTableQuery() string {
	return MergedCreateTable
}

func PurgeQuery() string {
	return purgeQuery
}

type SqlWrapper struct {
	event data.Event
	gnss  *gnss.GnssEvent
}

func NewSqlWrapper(event data.Event, gnss *gnss.GnssEvent) *SqlWrapper {
	return &SqlWrapper{
		event: event,
		gnss:  gnss,
	}
}

func (w *SqlWrapper) InsertQuery() (string, []any) {
	return insertQuery, []any{
		w.event.GetTime(),
		w.event.GetName(),
		w.gnss.Data.Latitude,
		w.gnss.Data.Longitude,
		w.gnss.Data.Speed,
	}
}
