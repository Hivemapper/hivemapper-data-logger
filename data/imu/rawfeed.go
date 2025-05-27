package imu

import (
	"fmt"
	"os"
	"time"

	"github.com/streamingfast/imu-controller/device/iim42652"
)

type RawFeed struct {
	imu                 *iim42652.IIM42652
	handlers            []RawFeedHandler
	fysnc_error_counter int
}

type ImuRawData struct {
	acceleration *Acceleration
	angularRate  *iim42652.AngularRate
	temperature  iim42652.Temperature
	fsync        *iim42652.Fsync
}

func NewRawFeed(imu *iim42652.IIM42652, handlers ...RawFeedHandler) *RawFeed {
	return &RawFeed{
		imu:      imu,
		handlers: handlers,
	}
}

type RawFeedHandler func(acceleration *Acceleration, angularRate *iim42652.AngularRate, temperature iim42652.Temperature, fsync *iim42652.Fsync) error

func (f *RawFeed) Run(axisMap *iim42652.AxisMap) error {
	fmt.Println("Run imu raw feed")

	// Open log file once before loop
	logFile, err := os.OpenFile("/data/logger_imu_loop.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer logFile.Close()

	dataChan := make(chan ImuRawData, 100) // buffered to absorb some load

	go func() {
		var count int
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case data := <-dataChan:
				for _, handler := range f.handlers {
					if err := handler(data.acceleration, data.angularRate, data.temperature, data.fsync); err != nil {
						fmt.Printf("handler error: %v\n", err)
					}
				}
				count++

			case <-ticker.C:
				fmt.Printf("Handler loop frequency: %d Hz\n", count)
				fmt.Printf("Buffered channel length: %d / %d\n", len(dataChan), cap(dataChan))
				count = 0
			}
		}
	}()

	for {
		fsync, err := f.imu.GetFsync()
		if err != nil {
			return fmt.Errorf("[ERROR] error getting fsync: %w", err)
		}
		// return early if fsync_int variable in is false
		if !fsync.FsyncInt {
			f.fysnc_error_counter++
			if f.fysnc_error_counter > 300000 {
				fmt.Println("[ERROR] 300,000 repeated fsync errors. Fsync is not being set.")
				f.fysnc_error_counter = 0
			}
			// time.Sleep(1 * time.Millisecond)
			continue
		}
		f.fysnc_error_counter = 0

		acceleration, err := f.imu.GetAcceleration()
		if err != nil {
			return fmt.Errorf("getting acceleration: %w", err)
		}

		angularRate, err := f.imu.GetGyroscopeData()
		if err != nil {
			return fmt.Errorf("getting angular rate: %w", err)
		}

		temperature, err := f.imu.GetTemperature()
		if err != nil {
			return fmt.Errorf("getting temperature: %w", err)
		}

		data := ImuRawData{
			acceleration: NewAcceleration(axisMap.X(acceleration), axisMap.Y(acceleration), axisMap.Z(acceleration), acceleration.TotalMagnitude, time.Now()),
			angularRate:  angularRate,
			temperature:  temperature,
			fsync:        fsync,
		}

		select {
		case dataChan <- data:
			// Sent successfully
		default:
			// Channel full, drop or log
			fmt.Println("Warning: data channel full, dropping data")
		}

		if angularRate.X < -2000.0 {
			fmt.Println("Resetting imu because angular rate is too high:", angularRate.X)
			err := f.imu.Init()
			if err != nil {
				return fmt.Errorf("initializing IMU: %w", err)
			}
		}
	}
}
