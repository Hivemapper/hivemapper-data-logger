package merged

import (
	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

const MergedCreateTable string = `
	CREATE TABLE IF NOT EXISTS merged (
		id INTEGER NOT NULL PRIMARY KEY,
		imu_time TIMESTAMP NOT NULL,
		imu_magnitude REAL NOT NULL,
		imu_acc_x REAL NOT NULL,
		imu_tilt_angle_x REAL NOT NULL,
		imu_acc_y REAL NOT NULL,
		imu_tilt_angle_y REAL NOT NULL,
		imu_acc_z REAL NOT NULL,
		imu_tilt_angle_z REAL NOT NULL,
		imu_temperature REAL NOT NULL,
		cam_orientation TEXT NOT NULL,
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
		gnss_rf_ofs_q INTEGER NOT NULL
	);
	create index if not exists merged_imu_time_idx on merged(imu_time);
`

const insertMergedQuery string = `INSERT INTO merged VALUES `
const insertMergedFields string = `(NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?),`

const purgeQuery string = `
	DELETE FROM merged WHERE imu_time < ?;
`

func CreateTableQuery() string {
	return MergedCreateTable
}

func PurgeQuery() string {
	return purgeQuery
}

type SqlWrapper struct {
	gnssData     *neom9n.Data
	acceleration *imu.Acceleration
	tiltAngles   *imu.TiltAngles
	temperature  iim42652.Temperature
	orientation  imu.Orientation
}

func NewSqlWrapper(acceleration *imu.Acceleration, tiltAngles *imu.TiltAngles, gnssData *neom9n.Data, temperature iim42652.Temperature, orientation imu.Orientation) *SqlWrapper {
	return &SqlWrapper{
		acceleration: acceleration,
		tiltAngles:   tiltAngles,
		orientation:  orientation,
		temperature:  temperature,
		gnssData:     gnssData,
	}
}

func (w *SqlWrapper) InsertQuery() (string, string, []any) {
	return insertMergedQuery, insertMergedFields, []any{
		w.acceleration.Time.Format("2006-01-02 15:04:05.99999"), //FIXME: remove the format only there for python a marde
		w.acceleration.Magnitude,
		w.acceleration.X,
		w.tiltAngles.X,
		w.acceleration.Y,
		w.tiltAngles.Y,
		w.acceleration.Z,
		w.tiltAngles.Z,
		*w.temperature,
		w.orientation,
		w.gnssData.SystemTime.Format("2006-01-02 15:04:05.99999"), //FIXME: remove the format only there for python a marde
		w.gnssData.Timestamp.Format("2006-01-02 15:04:05.99999"),  //FIXME: remove the format only there for python a marde
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
