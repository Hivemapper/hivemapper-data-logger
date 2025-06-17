package main

import (
	"fmt"
	"time"
	"sync"

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

	imuChan           chan *logger.ImuRedisWrapper
	gnssChan          chan *neom9n.Data
	magChan           chan *logger.MagnetometerRedisWrapper
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
	var imuChan chan *logger.ImuRedisWrapper
	var gnssChan chan *neom9n.Data
	var magChan chan *logger.MagnetometerRedisWrapper

	if redisLogsEnabled {
		redisLogger = logger.NewRedis(maxRedisImuEntries, maxRedisMagEntries, maxRedisGnssEntries, maxRedisGnssAuthEntries, redisLogProtoText)
		err := redisLogger.Init()
		if err != nil {
			return nil, fmt.Errorf("initializing redis logger database: %w", err)
		}

		imuChan = make(chan *logger.ImuRedisWrapper, 10000)
		gnssChan = make(chan *neom9n.Data, 10000)
		magChan = make(chan *logger.MagnetometerRedisWrapper, 10000)

		// IMU batch writer
		go func() {
			batch := make([]*logger.ImuRedisWrapper, 0, 100)
			ticker := time.NewTicker(100 * time.Millisecond)
			for {
				select {
				case msg := <-imuChan:
					batch = append(batch, msg)
					if len(batch) >= 50 {
						_ = redisLogger.LogImuDataBatch(batch)
						batch = batch[:0]
					}
				case <-ticker.C:
					if len(batch) > 0 {
						_ = redisLogger.LogImuDataBatch(batch)
						batch = batch[:0]
					}
				}
			}
		}()

		// GNSS batch writer
		go func() {
			batch := make([]*neom9n.Data, 0, 100)
			ticker := time.NewTicker(200 * time.Millisecond)
			for {
				select {
				case msg := <-gnssChan:
					batch = append(batch, msg)
					if len(batch) >= 50 {
						_ = redisLogger.LogGnssDataBatch(batch)
						batch = batch[:0]
					}
				case <-ticker.C:
					if len(batch) > 0 {
						_ = redisLogger.LogGnssDataBatch(batch)
						batch = batch[:0]
					}
				}
			}
		}()

		// Magnetometer batch writer
		go func() {
			batch := make([]*logger.MagnetometerRedisWrapper, 0, 100)
			ticker := time.NewTicker(200 * time.Millisecond)
			for {
				select {
				case msg := <-magChan:
					batch = append(batch, msg)
					if len(batch) >= 50 {
						_ = redisLogger.LogMagnetometerDataBatch(batch)
						batch = batch[:0]
					}
				case <-ticker.C:
					if len(batch) > 0 {
						_ = redisLogger.LogMagnetometerDataBatch(batch)
						batch = batch[:0]
					}
				}
			}
		}()
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
		imuChan:          imuChan,
		gnssChan:         gnssChan,
		magChan:          magChan,
	}, nil
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
	if h.redisLogsEnabled && h.gnssChan != nil {
		select {
		case h.gnssChan <- data:
		default:
			// drop
		}
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
	if h.redisLogsEnabled && h.magChan != nil {
		magData := logger.NewMagnetometerRedisWrapper(system_time, mag_x, mag_y, mag_z)
		select {
		case h.magChan <- magData:
		default:
			// drop
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

	if h.redisLogsEnabled && h.imuChan != nil {
		imuDataWrapper2 := logger.NewImuRedisWrapper(time.Now().UTC(), temperature, acceleration, angularRate)
		select {
		case h.imuChan <- imuDataWrapper2:
		default:
			// Optional: log dropped samples or increment a metric
		}
	}
	return nil
}
