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

type RawFeedHandler func(acceleration *Acceleration, angularRate *iim42652.AngularRate) error

func (f *RawFeed) Run() error {
	fmt.Println("Run imu raw feed")
	for {
		time.Sleep(10 * time.Millisecond)
		acceleration, err := f.imu.GetAcceleration()
		if err != nil {
			return fmt.Errorf("getting acceleration: %w", err)
		}
		angularRate, err := f.imu.GetGyroscopeData()

		fmt.Println("Sent raw imu event")
		for _, handler := range f.handlers {
			err := handler(
				NewAcceleration(acceleration.X, acceleration.Y, acceleration.Z, acceleration.TotalMagnitude, time.Now()),
				angularRate,
			)
			if err != nil {
				return fmt.Errorf("calling handler: %w", err)
			}
		}
	}

}
