package sensorredis

import (
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/Hivemapper/hivemapper-data-logger/data/session"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

const insertImuRawQuery string = `INSERT OR IGNORE INTO imu VALUES`

const insertImuRawFields string = `(NULL,?,?,?,?,?,?,?,?,?),`

type ImuSqlWrapper struct {
	acceleration *imu.Acceleration
	temperature  iim42652.Temperature
	gyroscope    *iim42652.AngularRate
}

func NewImuSqlWrapper(temperature iim42652.Temperature, acceleration *imu.Acceleration, gyroscope *iim42652.AngularRate) *ImuSqlWrapper {
	return &ImuSqlWrapper{
		acceleration: acceleration,
		temperature:  temperature,
		gyroscope:    gyroscope,
	}
}

func (w *ImuSqlWrapper) InsertQuery() (string, string, []any) {
	// very basic validation to prevent empty records on getting into database
	if w.acceleration == nil ||
		w.acceleration.Time.IsZero() {
		return "", "", nil
	}
	sessionID, err := session.GetSession()
	if err != nil {
		panic(err) // Handle error if any
	}

	return insertImuRawQuery, insertImuRawFields, []any{
		w.acceleration.Time.Format("2006-01-02 15:04:05.99999"),
		w.acceleration.X,
		w.acceleration.Y,
		w.acceleration.Z,
		w.gyroscope.X,
		w.gyroscope.Y,
		w.gyroscope.Z,
		*w.temperature,
		sessionID,
	}
}
