package imu

import (
	"fmt"
	"time"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type RawImuAccelerationEvent struct {
	*data.BaseEvent
	Acceleration *iim42652.Acceleration `json:"acceleration"`
}

func (e *RawImuAccelerationEvent) String() string {
	return "RawImuAccelerationEvent"
}

func NewRawImuAccelerationEvent(acc *iim42652.Acceleration) *RawImuAccelerationEvent {
	return &RawImuAccelerationEvent{
		BaseEvent:    data.NewBaseEvent("IMU_RAW_ACCELERATION_EVENT"),
		Acceleration: acc,
	}
}

type RawFeed struct {
	imu           *iim42652.IIM42652
	subscriptions data.Subscriptions
}

func NewRawFeed(imu *iim42652.IIM42652) *RawFeed {
	return &RawFeed{
		imu:           imu,
		subscriptions: make(data.Subscriptions),
	}
}

func (f *RawFeed) Subscribe(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	f.subscriptions[name] = sub
	return sub
}

func (f *RawFeed) Run() error {
	fmt.Println("Running pipeline")
	err := f.run()
	if err != nil {
		return fmt.Errorf("running pipeline: %w", err)
	}
	return nil
}

func (f *RawFeed) run() error {
	for {
		time.Sleep(10 * time.Millisecond)
		if len(f.subscriptions) == 0 {
			continue
		}
		acceleration, err := f.imu.GetAcceleration()
		if err != nil {
			panic(fmt.Errorf("getting acceleration: %w", err))
		}

		event := NewRawImuAccelerationEvent(acceleration)
		for _, subscription := range f.subscriptions {
			subscription.IncomingEvents <- event
		}
	}
}
