package sql

import (
	b64 "encoding/base64"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
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

const insertGnssAuthQuery string = `INSERT INTO gnss_auth VALUES`

const insertGnssAuthFields string = `(NULL,?,?,?,?,?,?),`

const gnssAuthPurgeQuery string = `
	DELETE FROM gnss_auth WHERE system_time < ?;
`

func GnssAuthCreateTableQuery() string {
	return GnssAuthCreateTable
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
	return insertGnssAuthQuery, insertGnssAuthFields, []any{
		w.gnssData.SecEcsignBuffer,
		w.gnssData.SecEcsign.MsgNum,
		b64.StdEncoding.EncodeToString(w.gnssData.SecEcsign.FinalHash[:]),
		b64.StdEncoding.EncodeToString(w.gnssData.SecEcsign.SessionId[:]),
		b64.StdEncoding.EncodeToString(w.gnssData.SecEcsign.EcdsaSignature[:]),
		time.Now(),
	}
}
