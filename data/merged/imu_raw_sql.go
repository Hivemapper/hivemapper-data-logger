package merged

import (
	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

const ImuRawCreateTable string = `
	CREATE TABLE IF NOT EXISTS imu_raw (
		id INTEGER NOT NULL PRIMARY KEY,
		imu_time TIMESTAMP NOT NULL,
		imu_acc_x REAL NOT NULL,
		imu_acc_y REAL NOT NULL,
		imu_acc_z REAL NOT NULL,
		imu_temperature REAL NOT NULL,
		gnss_system_time TIMESTAMP NOT NULL,
		gnss_time TIMESTAMP NOT NULL,
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
		gnss_rf_ofs_q INTEGER NOT NULL,
		image_file_name TEXT NOT NULL,
	);
`

const insertRawQuery string = `INSERT INTO imu_raw VALUES`

const insertRawFields string = `(NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),`

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
	acceleration      *imu.Acceleration
	temperature       iim42652.Temperature
	gnssData          *neom9n.Data
	lastImageFilename string
}

func NewImuRawSqlWrapper(temperature iim42652.Temperature, acceleration *imu.Acceleration, gnssData *neom9n.Data, lastImageFilename string) *ImuRawSqlWrapper {
	return &ImuRawSqlWrapper{
		acceleration:      acceleration,
		temperature:       temperature,
		gnssData:          gnssData,
		lastImageFilename: lastImageFilename,
	}
}

func (w *ImuRawSqlWrapper) InsertQuery() (string, string, []any) {
	return insertRawQuery, insertRawFields, []any{
		w.acceleration.Time.Format("2006-01-02 15:04:05.99999"),
		w.acceleration.Y, //this is not a mistake
		w.acceleration.Z, //this is not a mistake
		w.acceleration.X, //this is not a mistake
		*w.temperature,
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
		w.lastImageFilename,
	}
}
