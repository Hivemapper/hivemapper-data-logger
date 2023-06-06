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
	return fmt.Sprintf("RawImuAccelerationEvent %s", e.Acceleration.String())
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

func (f *RawFeed) Start() {
	fmt.Println("Running imu raw feed")
	go func() {
		err := f.run()
		if err != nil {
			panic(fmt.Errorf("running pipeline: %w", err))
		}
	}()
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
