package imu

import (
	"fmt"
	"time"

	"github.com/Hivemapper/hivemapper-data-logger/imu-controller/device/iim42652"
	"github.com/Hivemapper/hivemapper-data-logger/logger"
)

type Acceleration struct {
	X    float64
	Y    float64
	Z    float64
	Time time.Time
}

// AccelAxisMod modifies the axis of the acceleration data based on the camera type
func AccelAxisMod(x, y, z float64, time time.Time, camType logger.CamType) *Acceleration {
	if camType == logger.HDC {
		return &Acceleration{
			X:    z,
			Y:    x,
			Z:    y,
			Time: time,
		}
	} else if camType == logger.HDCS {
		return &Acceleration{
			X:    -y,
			Y:    x,
			Z:    z,
			Time: time,
		}
	} else {
		return &Acceleration{
			X:    x,
			Y:    y,
			Z:    z,
			Time: time,
		}
	}
}

// GyroAxisMod modifies the axis of the gyro data based on the camera type
func GyroAxisMod(m *iim42652.AngularRate, camType logger.CamType) *iim42652.AngularRate {
	if camType == logger.HDC {
		return &iim42652.AngularRate{
			RawX: m.RawZ,
			RawY: m.RawX,
			RawZ: m.RawY,
			X:    m.Z,
			Y:    m.X,
			Z:    m.Y,
		}
	} else if camType == logger.HDCS {
		return &iim42652.AngularRate{
			RawX: -m.RawY,
			RawY: m.RawX,
			RawZ: m.RawZ,
			X:    -m.Y,
			Y:    m.X,
			Z:    m.Z,
		}
	} else {
		return &iim42652.AngularRate{
			RawX: m.RawX,
			RawY: m.RawY,
			RawZ: m.RawZ,
			X:    m.X,
			Y:    m.Y,
			Z:    m.Z,
		}
	}
}

func (a *Acceleration) String() string {
	return fmt.Sprintf("Acceleration{x=%f, y=%f, z=%f, time=%s}", a.X, a.Y, a.Z, a.Time)
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
	fmt.Println("Device TYPE:", f.imu.DeviceType)

	for {
		// Check if data is ready
		var status byte = 0x00
		intStatus, err := f.imu.ReadRegister(iim42652.DATA_READY_INTERRUPT_STATUS)
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

			err = f.handler(
				AccelAxisMod(acceleration.X, acceleration.Y, acceleration.Z, time.Now(), f.imu.DeviceType),
				GyroAxisMod(angularRate, f.imu.DeviceType),
				temperature,
			)
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
			// Prevent hitting interrupt register too fast
			time.Sleep(25 * time.Microsecond)
		}
	}
}
