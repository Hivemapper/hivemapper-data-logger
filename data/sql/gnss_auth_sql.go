package sql

import (
	b64 "encoding/base64"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/session"
)

const GnssAuthCreateTable string = `
	CREATE TABLE IF NOT EXISTS gnss_auth (
		id INTEGER NOT NULL PRIMARY KEY,
		buffer TEXT NOT NULL,
		buffer_message_num INTEGER NOT NULL,
		buffer_hash TEXT NOT NULL,
		session_id TEXT NOT NULL,
		signature TEXT NOT NULL,
		system_time TIMESTAMP NOT NULL
	);
	create index if not exists gnss_time_idx on gnss_auth(system_time);
`

const GnssAuthAlterTable string = `
	ALTER TABLE gnss_auth ADD COLUMN session TEXT NOT NULL DEFAULT '';
`

const insertGnssAuthQuery string = `INSERT INTO gnss_auth VALUES`

const insertGnssAuthFields string = `(NULL,?,?,?,?,?,?,?),`

const gnssAuthPurgeQuery string = `
DELETE FROM gnss_auth WHERE rowid NOT IN (
	SELECT rowid FROM gnss_auth ORDER BY rowid DESC LIMIT 600
);
`

func GnssAuthCreateTableQuery() string {
	return GnssAuthCreateTable
}

func GnssAuthAlterTableQuery() string {
	return GnssAuthAlterTable
}

func GnssAuthPurgeQuery() string {
	return gnssAuthPurgeQuery
}

type GnssAuthSqlWrapper struct {
	gnssData *neom9n.Data
}

func NewGnssAuthSqlWrapper(gnssData *neom9n.Data) *GnssAuthSqlWrapper {
	return &GnssAuthSqlWrapper{
		gnssData: gnssData,
	}
}

func (w *GnssAuthSqlWrapper) InsertQuery() (string, string, []any) {
	sessionID, err := session.GetSession()
	if err != nil {
		panic(err) // Handle error if any
	}
	return insertGnssAuthQuery, insertGnssAuthFields, []any{
		w.gnssData.SecEcsignBuffer,
		w.gnssData.SecEcsign.MsgNum,
		b64.StdEncoding.EncodeToString(w.gnssData.SecEcsign.FinalHash[:]),
		b64.StdEncoding.EncodeToString(w.gnssData.SecEcsign.SessionId[:]),
		b64.StdEncoding.EncodeToString(w.gnssData.SecEcsign.EcdsaSignature[:]),
		time.Now().Format("2006-01-02 15:04:05.99999"),
		sessionID,
	}
}
