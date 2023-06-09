package merged

import (
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
)

const MergedCreateTable string = `
	CREATE TABLE IF NOT EXISTS merged (
		id INTEGER NOT NULL PRIMARY KEY,
		imu_time DATETIME NOT NULL,
		imu_total_magnitude REAL NOT NULL,
		imu_acc_x REAL NOT NULL,
		imu_corrected_acc_x REAL NOT NULL,
		imu_tilt_angle_x REAL NOT NULL,
		imu_acc_y REAL NOT NULL,
		imu_corrected_acc_y REAL NOT NULL,
		imu_tilt_angle_y REAL NOT NULL,
		imu_acc_z REAL NOT NULL,
		cam_orientation STRING NOT NULL,
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

const insertQuery string = `
	INSERT INTO merged VALUES(NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);
`

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
	imuRawEvent       *imu.RawImuEvent
	correctedImuEvent *imu.TiltCorrectedAccelerationEvent
	gnss              *gnss.GnssEvent
}

func NewSqlWrapper(imuRawEvent *imu.RawImuEvent, correctedImuEvent *imu.TiltCorrectedAccelerationEvent, gnss *gnss.GnssEvent) *SqlWrapper {
	return &SqlWrapper{
		imuRawEvent:       imuRawEvent,
		correctedImuEvent: correctedImuEvent,
		gnss:              gnss,
	}
}

func (w *SqlWrapper) InsertQuery() (string, []any) {
	return insertQuery, []any{
		w.imuRawEvent.Time,
		w.imuRawEvent.Acceleration.TotalMagnitude,
		w.imuRawEvent.Acceleration.CamX(), // -> imu_acc_x -> Z
		w.correctedImuEvent.X,
		w.correctedImuEvent.XAngle,
		w.imuRawEvent.Acceleration.CamY(), // -> imu_acc_y -> X
		w.correctedImuEvent.Y,
		w.correctedImuEvent.YAngle,
		w.imuRawEvent.Acceleration.CamZ(), // -> imu_acc_z -> Y
		w.correctedImuEvent.Orientation,
		w.gnss.Data.SystemTime,
		w.gnss.Data.Timestamp,
		w.gnss.Data.Fix,
		w.gnss.Data.Ttff,
		w.gnss.Data.Latitude,
		w.gnss.Data.Longitude,
		w.gnss.Data.Altitude,
		w.gnss.Data.Speed,
		w.gnss.Data.Heading,
		w.gnss.Data.Satellites.Seen,
		w.gnss.Data.Satellites.Used,
		w.gnss.Data.Eph,
		w.gnss.Data.HorizontalAccuracy,
		w.gnss.Data.VerticalAccuracy,
		w.gnss.Data.HeadingAccuracy,
		w.gnss.Data.SpeedAccuracy,
		w.gnss.Data.Dop.HDop,
		w.gnss.Data.Dop.VDop,
		w.gnss.Data.Dop.XDop,
		w.gnss.Data.Dop.YDop,
		w.gnss.Data.Dop.TDop,
		w.gnss.Data.Dop.PDop,
		w.gnss.Data.Dop.GDop,
		w.gnss.Data.RF.JammingState,
		w.gnss.Data.RF.AntStatus,
		w.gnss.Data.RF.AntPower,
		w.gnss.Data.RF.PostStatus,
		w.gnss.Data.RF.NoisePerMS,
		w.gnss.Data.RF.AgcCnt,
		w.gnss.Data.RF.JamInd,
		w.gnss.Data.RF.OfsI,
		w.gnss.Data.RF.MagI,
		w.gnss.Data.RF.OfsQ,
	}
}
