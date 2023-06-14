package merged

import (
	"database/sql"
	"fmt"

	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
)

const ImuRawCreateTable string = `
	CREATE TABLE IF NOT EXISTS imu_raw (
		id INTEGER NOT NULL PRIMARY KEY,
		imu_time DATETIME NOT NULL,
		imu_acc_x REAL NOT NULL,
		imu_acc_y REAL NOT NULL,
		imu_acc_z REAL NOT NULL,
		gnss_system_time DATETIME NOT NULL,
		gnss_time DATETIME NOT NULL,
		gnss_fix TEXT NOT NULL,
		gnss_ttff INTEGER NOT NULL,
		gnss_latitude REAL NOT NULL,
		gnss_longitude REAL NOT NULL,
		gnss_altitude REAL NOT NULL,
		gnss_speed REAL NOT NULL,
		gnss_heading REAL NOT NULL,
		gnss_satellites_seen INTEGER NOT NULL,
		gnss_satellites_used INTEGER NOT NULL,
		gnss_eph INTEGER NOT NULL,
		gnss_horizontal_accuracy REAL NOT NULL,
		gnss_vertical_accuracy	REAL NOT NULL,
		gnss_heading_accuracy REAL NOT NULL,
		gnss_speed_accuracy REAL NOT NULL,
		gnss_dop_h REAL NOT NULL,
		gnss_dop_v REAL NOT NULL,
		gnss_dop_x REAL NOT NULL,
		gnss_dop_y REAL NOT NULL,
		gnss_dop_t REAL NOT NULL,
		gnss_dop_p REAL NOT NULL,
		gnss_dop_g REAL NOT NULL,
		gnss_rf_jamming_state TEXT NOT NULL,
		gnss_rf_ant_status TEXT NOT NULL,
		gnss_rf_ant_power TEXT NOT NULL,
		gnss_rf_post_status INTEGER NOT NULL,
		gnss_rf_noise_per_ms INTEGER NOT NULL,
		gnss_rf_agc_cnt INTEGER NOT NULL,
		gnss_rf_jam_ind INTEGER NOT NULL,
		gnss_rf_ofs_i INTEGER NOT NULL,
		gnss_rf_mag_i INTEGER NOT NULL,
		gnss_rf_ofs_q INTEGER NOT NULL
	);
`

const insertRawQuery string = `INSERT INTO imu_raw VALUES`

const insertRawFields string = `(NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),`

const imuRawPurgeQuery string = `
	DELETE FROM imu_raw WHERE imu_time < ?;
`

func ImuRawCreateTableQuery() string {
	return ImuRawCreateTable
}

func ImuRawPurgeQuery() string {
	return imuRawPurgeQuery
}

type ImuRawSqlWrapper struct {
	acceleration *imu.Acceleration
	gnssData     *neom9n.Data
}

func NewImuRawSqlWrapper(acceleration *imu.Acceleration, gnssData *neom9n.Data) *ImuRawSqlWrapper {
	return &ImuRawSqlWrapper{
		acceleration: acceleration,
		gnssData:     gnssData,
	}
}

var imuRawPrepareStatement *sql.Stmt

func InitImuRaw(db *sql.DB) error {
	stmt, err := db.Prepare(insertRawQuery)
	if err != nil {
		return fmt.Errorf("preparing statement for inserting imu raw data: %w", err)
	}
	imuRawPrepareStatement = stmt
	return nil
}

func (w *ImuRawSqlWrapper) InsertQuery() (string, string, []any) {
	return insertRawQuery, insertRawFields, []any{
		w.acceleration.Time,
		w.acceleration.Y, //this is not a mistake
		w.acceleration.Z, //this is not a mistake
		w.acceleration.X, //this is not a mistake
		w.gnssData.SystemTime,
		w.gnssData.Timestamp,
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
	}
}
