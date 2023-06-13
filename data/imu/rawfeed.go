package imu

import (
	"fmt"
	"time"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type RawImuEvent struct {
	*data.BaseEvent
	Acceleration *Acceleration         `json:"acceleration"`
	AngularRate  *iim42652.AngularRate `json:"angular_rate"`
}

func (e *RawImuEvent) String() string {
	return fmt.Sprintf("RawImuEvent")
}

func NewRawImuEvent(acc *iim42652.Acceleration, angularRate *iim42652.AngularRate) *RawImuEvent {
	return &RawImuEvent{
		BaseEvent:    data.NewBaseEvent("IMU_RAW_ACCELERATION_EVENT", "IMU"),
		Acceleration: NewAcceleration(acc.CamX(), acc.CamY(), acc.CamZ(), acc.TotalMagnitude),
		AngularRate:  angularRate,
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
	fmt.Println("Starting imu raw feed")
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
		angularRate, err := f.imu.GetGyroscopeData()

		event := NewRawImuEvent(acceleration, angularRate)
		for _, subscription := range f.subscriptions {
			subscription.IncomingEvents <- event
		}
	}
}
