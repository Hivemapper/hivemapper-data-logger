package imu

import (
	"fmt"
	"runtime"
	"time"

	"github.com/streamingfast/imu-controller/device/iim42652"
)

type RawFeed struct {
	imu                 *iim42652.IIM42652
	handlers            []RawFeedHandler
	fysnc_error_counter int
}

func NewRawFeed(imu *iim42652.IIM42652, handlers ...RawFeedHandler) *RawFeed {
	return &RawFeed{
		imu:      imu,
		handlers: handlers,
	}
}

type RawFeedHandler func(acceleration *Acceleration, angularRate *iim42652.AngularRate, temperature iim42652.Temperature) error

// bToMb converts bytes to megabytes.
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func (f *RawFeed) Run(axisMap *iim42652.AxisMap) error {
	fmt.Println("Run imu raw feed")

	for {

		_, fsync_int, err := f.imu.GetFsync()
		if err != nil {
			return fmt.Errorf("[ERROR] error getting fsync: %w", err)
		}
		// return early if fsync_int variable in is false
		if !fsync_int {
			f.fysnc_error_counter++
			if f.fysnc_error_counter > 15000 {
				fmt.Println("[ERROR] 15,000 repeated fsync errors. Fsync is not being set.")
				f.fysnc_error_counter = 0
			}

			continue
		}
		f.fysnc_error_counter = 0

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("HeapAlloc: %v MiB, HeapSys: %v MiB, StackSys: %v MiB\n", bToMb(m.HeapAlloc), bToMb(m.HeapSys), bToMb(m.StackSys))
		fmt.Printf("NumGC: %d\n", m.NumGC)
		fmt.Println("Num Goroutines:", runtime.NumGoroutine())
		fmt.Printf("Sys: %v MiB\n", bToMb(m.Sys))
		fmt.Printf("HeapIdle: %v MiB, HeapReleased: %v MiB\n", bToMb(m.HeapIdle), bToMb(m.HeapReleased))

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

		for _, handler := range f.handlers {
			err := handler(
				NewAcceleration(axisMap.X(acceleration), axisMap.Y(acceleration), axisMap.Z(acceleration), acceleration.TotalMagnitude, time.Now()),
				angularRate,
				temperature,
			)
			if err != nil {
				return fmt.Errorf("calling handler: %w", err)
			}
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
