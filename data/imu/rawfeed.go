package imu

import (
	"fmt"
	"time"

	"github.com/streamingfast/imu-controller/device/iim42652"
)

type Acceleration struct {
	X         float64
	Y         float64
	Z         float64
	Magnitude float64
	Time      time.Time
}

func NewAcceleration(x, y, z, m float64, time time.Time) *Acceleration {
	return &Acceleration{
		X:         x,
		Y:         y,
		Z:         z,
		Magnitude: m,
		Time:      time,
	}
}

func (a *Acceleration) String() string {
	return fmt.Sprintf("Acceleration{x=%f, y=%f, z=%f, magnitude=%f, time=%s}", a.X, a.Y, a.Z, a.Magnitude, a.Time)
}

type RawFeed struct {
	imu     *iim42652.IIM42652
	handler RawFeedHandler
}

func NewRawFeed(imu *iim42652.IIM42652, handler RawFeedHandler) *RawFeed {
	return &RawFeed{
		imu:     imu,
		handler: handler,
	}
}

type RawFeedHandler func(acceleration *Acceleration, angularRate *iim42652.AngularRate, temperature iim42652.Temperature) error

func (f *RawFeed) Run() error {
	fmt.Println("Run imu raw feed")

	for {
		// time.Sleep(5 * time.Millisecond)
		var status byte = 0x00
		intStatus, err := f.imu.ReadRegister(iim42652.INT_STATUS)
		if err != nil {
			return fmt.Errorf("error reading register: %w", err)
		}
		status |= intStatus
		if status&iim42652.BIT_INT_STATUS_DRDY == iim42652.BIT_INT_STATUS_DRDY {
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

			err = f.handler(NewAcceleration(acceleration.X, acceleration.Y, acceleration.Z, acceleration.TotalMagnitude, time.Now()), angularRate, temperature)
			if err != nil {
				return fmt.Errorf("calling handler: %w", err)
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
}
