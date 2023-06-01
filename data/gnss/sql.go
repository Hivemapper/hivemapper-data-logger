package gnss

import (
	"database/sql"
	"fmt"

	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/logger"
)

const GnssCreateTable string = `
  CREATE TABLE IF NOT EXISTS gnss (
  	id INTEGER NOT NULL PRIMARY KEY,
  	time DATETIME NOT NULL,
  	system_time DATETIME NOT NULL,
	fix TEXT NOT NULL,
	Eph INTEGER NOT NULL,
	Sep INTEGER NOT NULL,
	latitude REAL NOT NULL,
	longitude REAL NOT NULL,
	altitude REAL NOT NULL,
	heading REAL NOT NULL,
	speed REAL NOT NULL,
	gdop REAL NOT NULL,
	hdop REAL NOT NULL,
	pdop REAL NOT NULL,
	tdop REAL NOT NULL,
	vdop REAL NOT NULL,
	xdop REAL NOT NULL,
	ydop REAL NOT NULL,
	seen INTEGER NOT NULL,
	used INTEGER NOT NULL,
	ttff INTEGER NOT NULL,
	rf_jamming_state STRING NOT NULL,
	rf_ant_status STRING NOT NULL,
	rf_ant_power STRING NOT NULL,
	rf_post_status INTEGER NOT NULL,
	rf_noise_per_ms INTEGER NOT NULL,
	rf_agc_cnt INTEGER NOT NULL,
	rf_jam_ind INTEGER NOT NULL,
	rf_ofsi INTEGER NOT NULL,
	rf_magif INTEGER NOT NULL,
	rf_ofsq INTEGER NOT NULL,
	rf_magq INTEGER NOT NULL
  );`

const insertQuery string = `
INSERT INTO gnss VALUES(NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);
`

const purgeQuery string = `
DELETE FROM gnss WHERE time < ?;
`

const lastPositionQuery string = `
SELECT latitude, longitude, altitude
FROM gnss
WHERE fix = '3D' or fix = '2D'
ORDER BY time DESC LIMIT 1;
`

type SqlWrapper struct {
	data *neom9n.Data
}

func NewSqlWrapper(data *neom9n.Data) *SqlWrapper {
	return &SqlWrapper{data: data}
}

func (s *SqlWrapper) InsertQuery() (string, []any) {
	return insertQuery, []any{
		s.data.Timestamp,
		s.data.SystemTime,
		s.data.Fix,
		s.data.Eph,
		s.data.Sep,
		s.data.Latitude,
		s.data.Longitude,
		s.data.Altitude,
		s.data.Heading,
		s.data.Speed,
		s.data.Dop.GDop,
		s.data.Dop.HDop,
		s.data.Dop.PDop,
		s.data.Dop.TDop,
		s.data.Dop.VDop,
		s.data.Dop.XDop,
		s.data.Dop.YDop,
		s.data.Satellites.Seen,
		s.data.Satellites.Used,
		s.data.Ttff,
		s.data.RF.JammingState,
		s.data.RF.AntStatus,
		s.data.RF.AntPower,
		s.data.RF.PostStatus,
		s.data.RF.NoisePerMS,
		s.data.RF.AgcCnt,
		s.data.RF.JamInd,
		s.data.RF.OfsI,
		s.data.RF.MagI,
		s.data.RF.OfsQ,
		s.data.RF.MagQ,
	}
}

func CreateTableQuery() string {
	return GnssCreateTable
}

func PurgeQuery() string {
	return purgeQuery
}

func GetLastPosition(sqlite *logger.Sqlite) (*neom9n.Position, error) {
	fmt.Println("getting last position")

	var position *neom9n.Position
	err := sqlite.SingleRowQuery(lastPositionQuery, func(rows *sql.Rows) error {
		position = &neom9n.Position{}
		err := rows.Scan(&position.Latitude, &position.Longitude, &position.Altitude)
		if err != nil {
			return fmt.Errorf("scanning last position: %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("querying last position: %s", err.Error())
	}

	return position, nil
}
