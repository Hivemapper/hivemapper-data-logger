package sql

import (
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/Hivemapper/hivemapper-data-logger/data/session"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

const ImuCreateTable string = `
	CREATE TABLE IF NOT EXISTS imu (
		id INTEGER NOT NULL PRIMARY KEY,
		time TIMESTAMP NOT NULL,
		acc_x REAL NOT NULL,
		acc_y REAL NOT NULL,
		acc_z REAL NOT NULL,
		gyro_x REAL NOT NULL,
		gyro_y REAL NOT NULL,
		gyro_z REAL NOT NULL,
		temperature REAL NOT NULL
	);
	create index if not exists imu_time_idx on imu(time);
`

const ImuAlterTable string = `
	ALTER TABLE imu ADD COLUMN session TEXT NOT NULL DEFAULT '';
`

const insertImuRawQuery string = `INSERT OR IGNORE INTO imu VALUES`

const insertImuRawFields string = `(NULL,?,?,?,?,?,?,?,?,?),`

const imuPurgeQuery string = `
DELETE FROM imu WHERE rowid NOT IN (
	SELECT rowid FROM imu ORDER BY rowid DESC LIMIT 150000
);
`

func ImuCreateTableQuery() string {
	return ImuCreateTable
}

func ImuAlterTableQuery() string {
	return ImuAlterTable
}

func ImuPurgeQuery() string {
	return imuPurgeQuery
}

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

func (w *ImuSqlWrapper) BufferSize() int {
	return 10;
}

func (w *ImuSqlWrapper) InsertQuery() (string, string, []any) {
	// very basic validation to prevent empty records on getting into database
	if w.acceleration == nil || 
		w.acceleration.Time.IsZero()  {
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
