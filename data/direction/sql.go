package direction

import (
	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
)

const MergedCreateTable string = `
  CREATE TABLE IF NOT EXISTS direction_events (
  	id INTEGER NOT NULL PRIMARY KEY,
	time TIMESTAMP NOT NULL,
	name TEXT NOT NULL,
	latitude REAL NOT NULL,
	longitude REAL NOT NULL,
	speed REAL NOT NULL
  );`

const insertDirectionEventsQuery string = `INSERT INTO direction_events VALUES `
const insertDirectionEventsFields string = `(NULL,?,?,?,?,?),`

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
	event    data.Event
	gnssData *neom9n.Data
}

func NewSqlWrapper(event data.Event, gnssData *neom9n.Data) *SqlWrapper {
	return &SqlWrapper{
		event:    event,
		gnssData: gnssData,
	}
}

func (w *SqlWrapper) InsertQuery() (string, string, []any) {
	return insertDirectionEventsQuery, insertDirectionEventsFields, []any{
		w.event.GetTime().Format("2006-01-02 15:04:05.99999"),
		w.event.GetName(),
		w.gnssData.Latitude,
		w.gnssData.Longitude,
		w.gnssData.Speed,
	}
}
