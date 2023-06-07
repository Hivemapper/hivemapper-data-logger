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
	XAngle float64 `json:"x_angle"`
	YAngle float64 `json:"Y_angle"`
}

func (e *CorrectedAccelerationEvent) String() string {
	return fmt.Sprintf("CorrectedAccelerationEvent: %f, %f, Angles x %f, y %f", e.X, e.Y, e.XAngle, e.YAngle)
}

func NewCorrectedAccelerationEvent(x, y, xAngle, yAngle float64) *CorrectedAccelerationEvent {
	return &CorrectedAccelerationEvent{
		BaseEvent: data.NewBaseEvent("IMU_RAW_ACCELERATION_EVENT"),
		X:         x,
		Y:         y,
		XAngle:    xAngle,
		YAngle:    yAngle,
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
				e := event.(*RawImuEvent)

				a := e.Acceleration
				ax := a.CamX()
				ay := a.CamY()
				az := a.CamZ()

				xAngle, yAngle := computeTiltAngles(ax, ay, az)
				correctedX, correctedY := computeCorrectedGForce(ax, ay, az)

				correctedEvent := NewCorrectedAccelerationEvent(correctedX, correctedY, xAngle, yAngle)
				for _, subscription := range f.subscriptions {
					subscription.IncomingEvents <- correctedEvent
				}
			}
		}
	}()
}
