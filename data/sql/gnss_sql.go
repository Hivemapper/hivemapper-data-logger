package sql

import (
	"encoding/json"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/session"
)

const GnssCreateTable string = `
	CREATE TABLE IF NOT EXISTS gnss (
		id INTEGER NOT NULL PRIMARY KEY,
		system_time TIMESTAMP NOT NULL,
		time TIMESTAMP NOT NULL,
		fix TEXT NOT NULL,
		ttff INTEGER NOT NULL,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		altitude REAL NOT NULL,
		speed REAL NOT NULL,
		heading REAL NOT NULL,
		satellites_seen INTEGER NOT NULL,
		satellites_used INTEGER NOT NULL,
		eph INTEGER NOT NULL,
		horizontal_accuracy REAL NOT NULL,
		vertical_accuracy	REAL NOT NULL,
		heading_accuracy REAL NOT NULL,
		speed_accuracy REAL NOT NULL,
		hdop REAL NOT NULL,
		vdop REAL NOT NULL,
		xdop REAL NOT NULL,
		ydop REAL NOT NULL,
		tdop REAL NOT NULL,
		pdop REAL NOT NULL,
		gdop REAL NOT NULL,
		rf_jamming_state TEXT NOT NULL,
		rf_ant_status TEXT NOT NULL,
		rf_ant_power TEXT NOT NULL,
		rf_post_status INTEGER NOT NULL,
		rf_noise_per_ms INTEGER NOT NULL,
		rf_agc_cnt INTEGER NOT NULL,
		rf_jam_ind INTEGER NOT NULL,
		rf_ofs_i INTEGER NOT NULL,
		rf_mag_i INTEGER NOT NULL,
		rf_ofs_q INTEGER NOT NULL,
		gga TEXT NOT NULL,
		rxm_measx TEXT NOT NULL
	);
	create index if not exists gnss_time_idx on gnss(system_time);
`

const insertGnssRawQuery string = `INSERT OR IGNORE INTO gnss (id, system_time, time, fix, ttff, latitude, longitude, altitude, speed, heading, satellites_seen, satellites_used, eph, horizontal_accuracy, vertical_accuracy, heading_accuracy, speed_accuracy, hdop, vdop, xdop, ydop, tdop, pdop, gdop, rf_jamming_state, rf_ant_status, rf_ant_power, rf_post_status, rf_noise_per_ms, rf_agc_cnt, rf_jam_ind, rf_ofs_i, rf_mag_i, rf_ofs_q, gga, rxm_measx, session, actual_system_time, unfiltered_latitude, unfiltered_longitude, time_resolved, cno) VALUES`

const insertGnssRawFields string = `(NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),`

const gnssPurgeQuery string = `
	DELETE FROM gnss WHERE rowid NOT IN (
		SELECT rowid FROM gnss ORDER BY rowid DESC LIMIT 60000
	);
`

func GnssCreateTableQuery() string {
	return GnssCreateTable
}

func GnssAlterTableQuerySessionUnfilteredAndResolved() string {
	return `
	ALTER TABLE gnss ADD COLUMN actual_system_time TIMESTAMP NOT NULL DEFAULT '0000-00-00 00:00:00';
	ALTER TABLE gnss ADD COLUMN unfiltered_latitude REAL NOT NULL DEFAULT 0;
	ALTER TABLE gnss ADD COLUMN unfiltered_longitude REAL NOT NULL DEFAULT 0;
	ALTER TABLE gnss ADD COLUMN time_resolved INTEGER NOT NULL DEFAULT 0;
`
}

func GnssAlterTableQueryCno() string {
	return `
	ALTER TABLE gnss ADD COLUMN cno REAL NOT NULL DEFAULT 0;
`
}

func GnssAlterTableQuerySession() string {
	return `
	ALTER TABLE gnss ADD COLUMN session TEXT NOT NULL DEFAULT '';
`
}

func GnssPurgeQuery() string {
	return gnssPurgeQuery
}

type GnssSqlWrapper struct {
	gnssData     *neom9n.Data
}

func NewGnssSqlWrapper(gnssData *neom9n.Data) *GnssSqlWrapper {
	return &GnssSqlWrapper{
		gnssData:     gnssData,
	}
}

func (w *GnssSqlWrapper) InsertQuery() (string, string, []any) {
	// very basic validation to prevent empty records on getting into database
	if w.gnssData == nil ||
		w.gnssData.SystemTime.IsZero() ||
		w.gnssData.Timestamp.IsZero() {
		return "", "", nil
	}

	rxmMeasx, err := json.Marshal(w.gnssData.RxmMeasx)
	if err != nil {
		return "", "", nil
	}
	sessionID, err := session.GetSession()
	if err != nil {
		panic(err) // Handle error if any
	}
	

	return insertGnssRawQuery, insertGnssRawFields, []interface{}{
		w.gnssData.SystemTime.Format("2006-01-02 15:04:05.99999"),
		w.gnssData.Timestamp.Format("2006-01-02 15:04:05.99999"),
		w.gnssData.Fix,
		w.gnssData.Ttff,
		w.gnssData.Latitude,
		w.gnssData.Longitude,
		w.gnssData.Altitude,
		w.gnssData.Speed,
		w.gnssData.Heading,
		w.gnssData.Satellites.Seen,
		w.gnssData.Satellites.Used,
		w.gnssData.Eph,
		w.gnssData.HorizontalAccuracy,
		w.gnssData.VerticalAccuracy,
		w.gnssData.HeadingAccuracy,
		w.gnssData.SpeedAccuracy,
		w.gnssData.Dop.HDop,
		w.gnssData.Dop.VDop,
		w.gnssData.Dop.XDop,
		w.gnssData.Dop.YDop,
		w.gnssData.Dop.TDop,
		w.gnssData.Dop.PDop,
		w.gnssData.Dop.GDop,
		w.gnssData.RF.JammingState,
		w.gnssData.RF.AntStatus,
		w.gnssData.RF.AntPower,
		w.gnssData.RF.PostStatus,
		w.gnssData.RF.NoisePerMS,
		w.gnssData.RF.AgcCnt,
		w.gnssData.RF.JamInd,
		w.gnssData.RF.OfsI,
		w.gnssData.RF.MagI,
		w.gnssData.RF.OfsQ,
		"",
		string(rxmMeasx),
		sessionID,
		w.gnssData.ActualSystemTime.Format("2006-01-02 15:04:05.99999"),
		0.0,
		0.0,
		w.gnssData.TimeResolved,
		w.gnssData.Cno,
	}
}

func (w *GnssSqlWrapper) BufferSize() int {
	return 10;
}
