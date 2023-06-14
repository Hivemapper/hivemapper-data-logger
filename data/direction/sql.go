package direction

import (
	"database/sql"
	"fmt"

	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
)

const MergedCreateTable string = `
  CREATE TABLE IF NOT EXISTS direction_events (
  	id INTEGER NOT NULL PRIMARY KEY,
	time DATETIME NOT NULL,
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

var prepareDirectionEventsStatement *sql.Stmt

func InitDirectionEvents(db *sql.DB) error {
	stmt, err := db.Prepare(insertDirectionEventsQuery)
	if err != nil {
		return fmt.Errorf("preparing statement for direction events: %w", err)
	}
	prepareDirectionEventsStatement = stmt
	return nil
}

func (w *SqlWrapper) InsertQuery() (string, string, []any) {
	return insertDirectionEventsQuery, insertDirectionEventsFields, []any{
		w.event.GetTime(),
		w.event.GetName(),
		w.gnssData.Latitude,
		w.gnssData.Longitude,
		w.gnssData.Speed,
	}
}
