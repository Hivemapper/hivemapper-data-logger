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
	fifoChan := make(chan ImuRawData, 100)

	go func() {
		var count int
		var fifoCount int
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

			case fifoChan := <-fifoChan:
				// Handle FIFO data if needed
				if fifoChan.acceleration == nil {
					fmt.Println("Received nil FIFO data, skipping")
				}
				fifoCount++

			case <-ticker.C:
				fmt.Printf("Handler loop frequency: %d Hz\n", count)
				fmt.Printf("Fifo loop frequency: %d Hz\n", fifoCount)
				// fmt.Printf("Register buffer length: %d / %d\n", len(dataChan), cap(dataChan))
				// fmt.Printf("FIFO buffer length: %d / %d\n", len(fifoChan), cap(fifoChan))
				count = 0
				fifoCount = 0
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
			time.Sleep(1 * time.Millisecond)
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

		fifopackets, err := f.imu.GetFifo() // Read FIFO data, if needed
		if err != nil {
			return fmt.Errorf("getting fifo data: %w", err)
		}

		data := ImuRawData{
			acceleration: NewAcceleration(axisMap.X(acceleration), axisMap.Y(acceleration), axisMap.Z(acceleration), acceleration.TotalMagnitude, time.Now()),
			angularRate:  angularRate,
			temperature:  temperature,
			fsync:        fsync,
		}
		// fmt.Println("Register fsync data:", fsync)
		// fmt.Println("Register acceleration data:", acceleration)
		// fmt.Println("Register angular rate data:", angularRate)
		// fmt.Printf("Register temperature %.2f°C\n", *temperature)

		// print all fifo data objects
		for _, fifoData := range fifopackets {
			// fmt.Println("FIFO Fsync:", fifoData.Fsync)
			// fmt.Println("FIFO Acceleration:", fifoData.Acceleration)
			// fmt.Println("FIFO Angular Rate:", fifoData.AngularRate)
			// fmt.Printf("FIFO Temperature: %.2f°C\n", *fifoData.Temperature)
			fifo_raw_data := ImuRawData{
				acceleration: NewAcceleration(axisMap.X(fifoData.Acceleration), axisMap.Y(fifoData.Acceleration), axisMap.Z(fifoData.Acceleration), fifoData.Acceleration.TotalMagnitude, time.Now()),
				angularRate:  fifoData.AngularRate,
				temperature:  fifoData.Temperature,
				fsync:        fifoData.Fsync,
			}
			select {
			case fifoChan <- fifo_raw_data:
				// Sent successfully
			default:
				// Channel full, drop or log
				fmt.Println("Warning: fifo data channel full, dropping FIFO data")
			}
		}

		select {
		case dataChan <- data:
			// Sent successfully
		default:
			// Channel full, drop or log
			fmt.Println("Warning: register data channel full, dropping data")
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
