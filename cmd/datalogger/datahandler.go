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
	gnssAuthCount     int
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
		[]logger.CreateTableQueryFunc{sql.GnssCreateTableQuery, sql.GnssAuthCreateTableQuery, sql.ImuCreateTableQuery, sql.ErrorLogsCreateTableQuery, magnetometer.CreateTableQuery},
		[]logger.PurgeQueryFunc{sql.GnssPurgeQuery, sql.GnssAuthPurgeQuery, sql.ImuPurgeQuery, sql.ErrorLogsPurgeQuery, magnetometer.PurgeQuery})
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
		if h.gnssAuthCount%60 == 0 {
			err := h.sqliteLogger.Log(sql.NewGnssAuthSqlWrapper(data))
			if err != nil {
				return fmt.Errorf("logging gnss auth data to sqlite: %w", err)
			}
		}
		h.gnssAuthCount += 1
		return nil
	}

	return nil
}

func calibrate(mag_x float64, mag_y float64, mag_z float64, transform [3][3]float64, center [3]float64) [3]float64 {
	mag := [3]float64{mag_x, mag_y, mag_z}
	for i := 0; i < 3; i++ {
		mag[i] -= center[i]
	}

	results := [3]float64{0, 0, 0}
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			results[row] += transform[row][col] * mag[col]
		}
	}
	return results
}

func (h *DataHandler) HandlerMagnetometerData(system_time time.Time, mag_x float64, mag_y float64, mag_z float64) error {
	var center [3]float64
	var transform [3][3]float64
	calibrationString := h.sqliteLogger.GetConfig("magnetometerCalibration")
	_, err := fmt.Sscanf(calibrationString, "%f %f %f %f %f %f %f %f %f %f %f %f",
		&transform[0][0], &transform[0][1], &transform[0][2],
		&transform[1][0], &transform[1][1], &transform[1][2],
		&transform[2][0], &transform[2][1], &transform[2][2],
		&center[0], &center[1], &center[2])
	if err != nil {
		center = [3]float64{0, 0, 0}
		transform = [3][3]float64{
			{1, 0, 0},
			{0, 1, 0},
			{0, 0, 1},
		}
	}

	calibrated_mag := calibrate(mag_x, mag_y, mag_z, transform, center)
	err = h.sqliteLogger.Log(magnetometer.NewMagnetometerSqlWrapper(system_time, calibrated_mag[0], calibrated_mag[1], calibrated_mag[2]))
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
