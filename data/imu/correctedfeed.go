package imu

import (
	"fmt"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type CorrectedAccelerationEvent struct {
	*data.BaseEvent
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Z      float64 `json:"z"`
	XAngle float64 `json:"x_angle"`
	YAngle float64 `json:"Y_angle"`
	ZAngle float64 `json:"Z_angle"`
}

func (e *CorrectedAccelerationEvent) String() string {
	return fmt.Sprintf("CorrectedAccelerationEvent: %f, %f, %f Angles x %f, y %f, z %f", e.X, e.Y, e.Z, e.XAngle, e.YAngle, e.ZAngle)
}

func NewCorrectedAccelerationEvent(x, y, z, xAngle, yAngle, zAngle float64) *CorrectedAccelerationEvent {
	return &CorrectedAccelerationEvent{
		BaseEvent: data.NewBaseEvent("IMU_RAW_ACCELERATION_EVENT"),
		X:         x,
		Y:         y,
		Z:         z,
		XAngle:    xAngle,
		YAngle:    yAngle,
		ZAngle:    zAngle,
	}
}

type CorrectedAccelerationFeed struct {
	imu           *iim42652.IIM42652
	subscriptions data.Subscriptions
}

func NewCorrectedAccelerationFeed() *CorrectedAccelerationFeed {
	return &CorrectedAccelerationFeed{
		subscriptions: make(data.Subscriptions),
	}
}

func (f *CorrectedAccelerationFeed) Subscribe(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	f.subscriptions[name] = sub
	return sub
}

func (f *CorrectedAccelerationFeed) Start(raw *RawFeed) {
	fmt.Println("Running imu corrected feed")
	sub := raw.Subscribe("corrected")
	go func() {
		for {
			select {
			case event := <-sub.IncomingEvents:
				if len(f.subscriptions) == 0 {
					continue
				}

				e := event.(*RawImuAccelerationEvent)
				a := e.Acceleration
				x := a.CamX()
				y := a.CamY()
				z := a.CamZ()

				//todo: need to compute the corrected values from the x,y,z values at once
				// also need to understand how the x, y, z angles works to properly understand
				// how to calculate the GForce

				// todo: seems that we need to fetch some gyro data, to know about the tilt of the cam over time...
				correctedXAngle, correctedYAngle, correctedZAngle := computeCorrectedTiltAngles(x, y, z)
				correctedX, correctedY, correctedZ := computeCorrectedGForce(x, y, z, correctedXAngle, correctedYAngle, correctedZAngle)

				correctedEvent := NewCorrectedAccelerationEvent(correctedX, correctedY, correctedZ, correctedXAngle, correctedYAngle, correctedZAngle)
				for _, subscription := range f.subscriptions {
					subscription.IncomingEvents <- correctedEvent
				}
			}
		}
	}()
}
