package main

import (
	"fmt"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/Hivemapper/hivemapper-data-logger/logger"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type DataHandler struct {
	redisLogger       *logger.Redis
	gnssJsonLogger    *logger.JsonFile
	imuJsonLogger     *logger.JsonFile
	gnssData          *neom9n.Data
	lastImageFileName string
	jsonLogsEnabled   bool
	redisLogsEnabled  bool
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
	redisLogsEnabled bool,
	maxRedisImuEntries int,
	maxRedisMagEntries int,
	maxRedisGnssEntries int,
	maxRedisGnssAuthEntries int,
	redisLogProtoText bool,
) (*DataHandler, error) {

	var redisLogger *logger.Redis = nil
	if redisLogsEnabled {
		redisLogger = logger.NewRedis(maxRedisImuEntries, maxRedisMagEntries, maxRedisGnssEntries, maxRedisGnssAuthEntries, redisLogProtoText)
		err := redisLogger.Init()
		if err != nil {
			return nil, fmt.Errorf("initializing redis logger database: %w", err)
		}
	}

	gnssJsonLogger := logger.NewJsonFile(gnssJsonDestFolder, gnssSaveInterval)
	err := gnssJsonLogger.Init(false)
	if err != nil {
		return nil, fmt.Errorf("initializing gnss json logger: %w", err)
	}

	imuJsonLogger := logger.NewJsonFile(imuJsonDestFolder, imuSaveInterval)
	err = imuJsonLogger.Init(jsonLogsEnabled)
	if err != nil {
		return nil, fmt.Errorf("initializing imu json logger: %w", err)
	}

	return &DataHandler{
		redisLogger:      redisLogger,
		gnssJsonLogger:   gnssJsonLogger,
		imuJsonLogger:    imuJsonLogger,
		jsonLogsEnabled:  jsonLogsEnabled,
		redisLogsEnabled: redisLogsEnabled,
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
	return nil
}

func (h *DataHandler) HandlerGnssData(data *neom9n.Data) error {
	if data.SecEcsign == nil {
		h.gnssData = data
		if h.jsonLogsEnabled && !h.gnssJsonLogger.IsLogging && data.Fix != "none" {
			h.gnssJsonLogger.StartStoring()
		}
		err := h.gnssJsonLogger.Log(data.Timestamp, data)
		if err != nil {
			return fmt.Errorf("logging gnss data to json: %w", err)
		}

		if h.redisLogsEnabled {
			err = h.redisLogger.LogGnssData(*data)
			if err != nil {
				return fmt.Errorf("logging gnss data to redis: %w", err)
			}
		}
	} else {
		if h.gnssAuthCount%60 == 0 {
			if h.redisLogsEnabled {
				err := h.redisLogger.LogGnssAuthData(*data)
				if err != nil {
					return fmt.Errorf("logging gnss data to redis: %w", err)
				}
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
	center = [3]float64{0, 0, 0}
	transform = [3][3]float64{
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	}

	calibrated_mag := calibrate(mag_x, mag_y, mag_z, transform, center)
	magDataWrapper := logger.NewMagnetometerRedisWrapper(system_time, calibrated_mag[0], calibrated_mag[1], calibrated_mag[2])
	if h.redisLogsEnabled {
		err := h.redisLogger.LogMagnetometerData(*magDataWrapper)
		if err != nil {
			return fmt.Errorf("logging magnetometer data to redis: %w", err)
		}
	}

	return nil
}

func (h *DataHandler) HandleRawImuFeed(acceleration *imu.Acceleration, angularRate *iim42652.AngularRate, temperature iim42652.Temperature) error {
	imuDataWrapper := logger.NewImuDataWrapper(temperature, acceleration, angularRate)
	err := h.imuJsonLogger.Log(time.Now().UTC(), imuDataWrapper)
	if err != nil {
		return fmt.Errorf("logging raw imu data to json: %w", err)
	}

	imuDataWrapper2 := logger.NewImuRedisWrapper(time.Now().UTC(), temperature, acceleration, angularRate)
	if h.redisLogsEnabled {
		err = h.redisLogger.LogImuData(*imuDataWrapper2)
		if err != nil {
			return fmt.Errorf("logging raw imu data to redis: %w", err)
		}
	}
	return nil
}