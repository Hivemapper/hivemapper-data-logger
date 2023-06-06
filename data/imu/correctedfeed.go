package imu

import (
	"fmt"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type CorrectedAccelerationEvent struct {
	*data.BaseEvent
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func (e *CorrectedAccelerationEvent) String() string {
	return fmt.Sprintf("CorrectedAccelerationEvent: %f, %f, %f", e.X, e.Y, e.Z)
}

func NewCorrectedAccelerationEvent(x, y, z float64) *CorrectedAccelerationEvent {
	return &CorrectedAccelerationEvent{
		BaseEvent: data.NewBaseEvent("IMU_RAW_ACCELERATION_EVENT"),
		X:         x,
		Y:         y,
		Z:         z,
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

func (f *CorrectedAccelerationFeed) Run(raw *RawFeed) error {
	fmt.Println("Running imu corrected feed")
	sub := raw.Subscribe("corrected")
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
			correctedX := computeCorrectedGForce(z, x)
			correctedY := computeCorrectedGForce(z, y)
			correctedZ := z

			correctedEvent := NewCorrectedAccelerationEvent(correctedX, correctedY, correctedZ)
			for _, subscription := range f.subscriptions {
				subscription.IncomingEvents <- correctedEvent
			}

		}
	}
}
