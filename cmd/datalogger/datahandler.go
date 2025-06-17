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

const (
	BatchSize = 100
)

type BatchHandler struct {
	imuBuffer     []*imu.Acceleration
	gnssBuffer    []*neom9n.Data
	imuMutex      sync.Mutex
	gnssMutex     sync.Mutex
	lastImuFlush  time.Time
	lastGnssFlush time.Time
	flushInterval time.Duration
	redisLogsEnabled bool
	jsonLogsEnabled  bool
	gnssAuthCount     int
	imuJsonLogger     *logger.JsonFile
	gnssJsonLogger    *logger.JsonFile
	redisLogger       *logger.Redis
}

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
	batchHandler      *BatchHandler
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

	batchHandler := &BatchHandler{
		imuBuffer:     make([]*imu.Acceleration, 0, BatchSize),
		gnssBuffer:    make([]*neom9n.Data, 0, BatchSize),
		imuMutex:      sync.Mutex{},
		gnssMutex:     sync.Mutex{},
		lastImuFlush:  time.Time{},
		lastGnssFlush: time.Time{},
		flushInterval: 100 * time.Millisecond,
		redisLogsEnabled: redisLogsEnabled,
		jsonLogsEnabled:  jsonLogsEnabled,
		gnssAuthCount:     0,
		imuJsonLogger:     imuJsonLogger,
		gnssJsonLogger:    gnssJsonLogger,
		redisLogger:       redisLogger,
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
		batchHandler:      batchHandler,
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
	return h.batchHandler.AddImuData(acceleration, angularRate, temperature)
}

func (b *BatchHandler) AddImuData(acceleration *imu.Acceleration, angularRate *iim42652.AngularRate, temperature iim42652.Temperature) error {
	b.imuMutex.Lock()
	defer b.imuMutex.Unlock()

	// Store all IMU data together
	b.imuBuffer = append(b.imuBuffer, acceleration)
	if len(b.imuBuffer) >= BatchSize || time.Since(b.lastImuFlush) >= b.flushInterval {
		return b.flushImu(angularRate, temperature)
	}
	return nil
}

func (b *BatchHandler) flushImu(angularRate *iim42652.AngularRate, temperature iim42652.Temperature) error {
	b.imuMutex.Lock()
	defer b.imuMutex.Unlock()

	if len(b.imuBuffer) == 0 {
		return nil
	}

	// Process batch
	for _, data := range b.imuBuffer {
		imuDataWrapper := logger.NewImuDataWrapper(temperature, data, angularRate)
		err := b.imuJsonLogger.Log(time.Now().UTC(), imuDataWrapper)
		if err != nil {
			return fmt.Errorf("batch logging raw imu data to json: %w", err)
		}

		if b.redisLogsEnabled {
			imuDataWrapper2 := logger.NewImuRedisWrapper(time.Now().UTC(), temperature, data, angularRate)
			err = b.redisLogger.LogImuData(*imuDataWrapper2)
			if err != nil {
				return fmt.Errorf("batch logging raw imu data to redis: %w", err)
			}
		}
	}

	b.imuBuffer = b.imuBuffer[:0]
	b.lastImuFlush = time.Now()
	return nil
}

func (b *BatchHandler) flushGnss() error {
	b.gnssMutex.Lock()
	defer b.gnssMutex.Unlock()

	if len(b.gnssBuffer) == 0 {
		return nil
	}

	// Process batch
	for _, data := range b.gnssBuffer {
		if data.SecEcsign == nil {
			if b.jsonLogsEnabled && !b.gnssJsonLogger.IsLogging && data.Fix != "none" {
				b.gnssJsonLogger.StartStoring()
			}
			err := b.gnssJsonLogger.Log(data.Timestamp, data)
			if err != nil {
				return fmt.Errorf("batch logging gnss data to json: %w", err)
			}

			if b.redisLogsEnabled {
				err = b.redisLogger.LogGnssDataBatch([]*neom9n.Data{data})
				if err != nil {
					return fmt.Errorf("batch logging gnss data to redis: %w", err)
				}
			}
		} else {
			if b.gnssAuthCount%60 == 0 {
				if b.redisLogsEnabled {
					err := b.redisLogger.LogGnssDataBatch([]*neom9n.Data{data})
					if err != nil {
						return fmt.Errorf("batch logging gnss auth data to redis: %w", err)
					}
				}
			}
			b.gnssAuthCount += 1
		}
	}

	b.gnssBuffer = b.gnssBuffer[:0]
	b.lastGnssFlush = time.Now()
	return nil
}
