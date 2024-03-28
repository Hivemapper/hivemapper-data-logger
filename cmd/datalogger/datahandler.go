package main

import (
	"fmt"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data"
	"github.com/Hivemapper/hivemapper-data-logger/data/direction"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/Hivemapper/hivemapper-data-logger/data/magnetometer"
	"github.com/Hivemapper/hivemapper-data-logger/data/sql"
	"github.com/Hivemapper/hivemapper-data-logger/logger"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type DataHandler struct {
	sqliteLogger      *logger.Sqlite
	gnssJsonLogger    *logger.JsonFile
	imuJsonLogger     *logger.JsonFile
	gnssData          *neom9n.Data
	lastImageFileName string
	jsonLogsEnabled   bool
}

func NewDataHandler(
	dbPath string,
	dbLogTTL time.Duration,
	gnssJsonDestFolder string,
	gnssSaveInterval time.Duration,
	imuJsonDestFolder string,
	imuSaveInterval time.Duration,
	jsonLogsEnabled bool,
) (*DataHandler, error) {
	sqliteLogger := logger.NewSqlite(
		dbPath,
		[]logger.CreateTableQueryFunc{sql.GnssCreateTableQuery, sql.GnssAuthCreateTableQuery, sql.ImuCreateTableQuery, sql.ErrorLogsCreateTableQuery, direction.CreateTableQuery, magnetometer.CreateTableQuery},
		[]logger.PurgeQueryFunc{sql.GnssPurgeQuery, sql.GnssAuthPurgeQuery, sql.ImuPurgeQuery, sql.ErrorLogsPurgeQuery, direction.PurgeQuery, magnetometer.PurgeQuery})
	err := sqliteLogger.Init(dbLogTTL)
	if err != nil {
		return nil, fmt.Errorf("initializing sqlite logger database: %w", err)
	}

	gnssJsonLogger := logger.NewJsonFile(gnssJsonDestFolder, gnssSaveInterval)
	err = gnssJsonLogger.Init(false)
	if err != nil {
		return nil, fmt.Errorf("initializing gnss json logger: %w", err)
	}

	imuJsonLogger := logger.NewJsonFile(imuJsonDestFolder, imuSaveInterval)
	err = imuJsonLogger.Init(jsonLogsEnabled)
	if err != nil {
		return nil, fmt.Errorf("initializing imu json logger: %w", err)
	}

	return &DataHandler{
		sqliteLogger:    sqliteLogger,
		gnssJsonLogger:  gnssJsonLogger,
		imuJsonLogger:   imuJsonLogger,
		jsonLogsEnabled: jsonLogsEnabled,
	}, err
}

func (h *DataHandler) HandleImage(imageFileName string) error {
	h.lastImageFileName = imageFileName
	return nil
}

func (h *DataHandler) HandleOrientedAcceleration(
	acceleration *imu.Acceleration,
	tiltAngles *imu.TiltAngles,
	temperature iim42652.Temperature,
	orientation imu.Orientation,
) error {
	// gnssData := mustGnssEvent(h.gnssData)
	// err := h.sqliteLogger.Log(merged.NewSqlWrapper(acceleration, tiltAngles, gnssData, temperature, orientation))
	// if err != nil {
	// 	return fmt.Errorf("logging merged data to sqlite: %w", err)
	// }
	return nil
}

func (h *DataHandler) HandlerGnssData(data *neom9n.Data) error {
	if data.SecEcsign == nil {
		h.gnssData = data
		if h.jsonLogsEnabled && !h.gnssJsonLogger.IsLogging && data.Fix != "none" {
			h.gnssJsonLogger.StartStoring()
		}
		err := h.sqliteLogger.Log(sql.NewGnssSqlWrapper(data))
		if err != nil {
			return fmt.Errorf("logging raw gnss data to sqlite: %w", err)
		}
		err = h.gnssJsonLogger.Log(data.Timestamp, data)
		if err != nil {
			return fmt.Errorf("logging gnss data to json: %w", err)
		}
	} else {
		err := h.sqliteLogger.Log(sql.NewGnssAuthSqlWrapper(data))
		if err != nil {
			return fmt.Errorf("logging gnss auth data to sqlite: %w", err)
		}
		return nil
	}

	return nil
}

func (h *DataHandler) HandlerMagnetometerData(system_time time.Time, mag_x float64, mag_y float64, mag_z float64) error {
	err := h.sqliteLogger.Log(magnetometer.NewMagnetometerSqlWrapper(system_time, mag_x, mag_y, mag_z))
	if err != nil {
		return fmt.Errorf("logging magnetometer data to sqlite: %w", err)
	}
	return nil
}

func (h *DataHandler) HandleRawImuFeed(acceleration *imu.Acceleration, angularRate *iim42652.AngularRate, temperature iim42652.Temperature) error {
	err := h.sqliteLogger.Log(sql.NewImuSqlWrapper(temperature, acceleration, angularRate))
	if err != nil {
		return fmt.Errorf("logging raw imu data to sqlite: %w", err)
	}
	imuDataWrapper := logger.NewImuDataWrapper(temperature, acceleration, angularRate)
	err = h.imuJsonLogger.Log(time.Now(), imuDataWrapper)
	if err != nil {
		return fmt.Errorf("logging raw imu data to json: %w", err)
	}
	return nil
}

func (h *DataHandler) HandleDirectionEvent(event data.Event) error {
	gnssData := mustGnssEvent(h.gnssData)
	err := h.sqliteLogger.Log(direction.NewSqlWrapper(event, gnssData))
	if err != nil {
		return fmt.Errorf("logging direction data to sqlite: %w", err)
	}
	return nil
}
