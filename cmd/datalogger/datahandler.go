package main

import (
	"fmt"
	"time"

	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/Hivemapper/hivemapper-data-logger/data/sql"
	"github.com/Hivemapper/hivemapper-data-logger/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/imu-controller/device/iim42652"
	"github.com/Hivemapper/hivemapper-data-logger/logger"
)

type DataHandler struct {
	sqliteLogger  *logger.Sqlite
	gnssData      *neom9n.Data
	gnssAuthCount int
}

func NewDataHandler(
	dbPath string,
) (*DataHandler, error) {
	sqliteLogger := logger.NewSqlite(
		dbPath,
		[]logger.CreateTableQueryFunc{sql.GnssCreateTableQuery, sql.GnssAuthCreateTableQuery, sql.ImuCreateTableQuery, sql.MagCreateTableQuery},
		[]logger.AlterTableQueryFunc{sql.GnssAlterTableQuerySession, sql.GnssAlterTableQuerySessionUnfilteredAndResolved, sql.GnssAuthAlterTableQuery, sql.ImuAlterTableQuery, sql.MagAlterTableQuery})
	err := sqliteLogger.Init()
	if err != nil {
		return nil, fmt.Errorf("initializing sqlite logger database: %w", err)
	}

	return &DataHandler{
		sqliteLogger: sqliteLogger,
	}, err
}

func (h *DataHandler) HandlerGnssData(data *neom9n.Data) error {
	if data.SecEcsign == nil {
		h.gnssData = data
		err := h.sqliteLogger.Log(sql.NewGnssSqlWrapper(data))
		if err != nil {
			return fmt.Errorf("logging raw gnss data to sqlite: %w", err)
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
	err = h.sqliteLogger.Log(sql.NewMagnetometerSqlWrapper(system_time, calibrated_mag[0], calibrated_mag[1], calibrated_mag[2]))
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
	return nil
}
