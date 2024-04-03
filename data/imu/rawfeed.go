package imu

import (
	"fmt"
	"time"

	"github.com/streamingfast/imu-controller/device/iim42652"
)

type RawFeed struct {
	imu      *iim42652.IIM42652
	handlers []RawFeedHandler
}

func NewRawFeed(imu *iim42652.IIM42652, handlers ...RawFeedHandler) *RawFeed {
	return &RawFeed{
		imu:      imu,
		handlers: handlers,
	}
}

type RawFeedHandler func(acceleration *Acceleration, angularRate *iim42652.AngularRate, temperature iim42652.Temperature) error

//TODO: add FileWatcherEventFeed
// and have imu raw subscribe to it
// have 1 jpg that we will keep on reusing the same image
// then create the frameKms and the gz files with the same image over and over again
// inspire on the file watcher in the hdc-debugger

func (f *RawFeed) Run(axisMap *iim42652.AxisMap) error {
    fmt.Println("Run imu raw feed")

    var count int
    var totalAccelX, totalAccelY, totalAccelZ, totalAccelMag, totalAngularX, totalAngularY, totalAngularZ, totalTemperature float64
	windowSize := 5
    for {
        time.Sleep(20 * time.Millisecond)
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

        // Accumulate values
        totalAccelX += axisMap.X(acceleration)
        totalAccelY += axisMap.Y(acceleration)
        totalAccelZ += axisMap.Z(acceleration)
        totalAccelMag += acceleration.TotalMagnitude
        totalAngularX += angularRate.X
        totalAngularY += angularRate.Y
        totalAngularZ += angularRate.Z
        totalTemperature += float64(*temperature)
        count++

        if count == windowSize {
            // Compute averages
            size := float64(windowSize)
            avgAccelX := totalAccelX / size
            avgAccelY := totalAccelY / size
            avgAccelZ := totalAccelZ / size
            avgAccelMag := totalAccelMag / size
            avgAngularX := totalAngularX / size
            avgAngularY := totalAngularY / size
            avgAngularZ := totalAngularZ / size
            avgTemperature := totalTemperature / size

            // Call handlers with averages
            for _, handler := range f.handlers {
                err := handler(
                    NewAcceleration(avgAccelX, avgAccelY, avgAccelZ, avgAccelMag, time.Now()),
                    &iim42652.AngularRate{int16(avgAngularX), int16(avgAngularY), int16(avgAngularZ), avgAngularX, avgAngularY, avgAngularZ},
                    iim42652.NewTemperature(avgTemperature),
                )
                if err != nil {
                    return fmt.Errorf("calling handler: %w", err)
                }
            }

            // Reset accumulators and counter
            totalAccelX, totalAccelY, totalAccelZ, totalAccelMag, totalAngularX, totalAngularY, totalAngularZ, totalTemperature = 0, 0, 0, 0, 0, 0, 0, 0
            count = 0
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
