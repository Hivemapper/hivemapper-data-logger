package direction

import (
	"fmt"

	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
)

type DirectionEventHandler func(event data.Event) error

type DirectionEventFeed struct {
	config              *imu.Config
	leftTurnTracker     *LeftTurnTracker
	rightTurnTracker    *RightTurnTracker
	accelerationTracker *AccelerationTracker
	decelerationTracker *DecelerationTracker
	stopTracker         *StopTracker

	gnssData *neom9n.Data
	handlers []DirectionEventHandler
}

func NewDirectionEventFeed(config *imu.Config, handlers ...DirectionEventHandler) *DirectionEventFeed {
	feed := &DirectionEventFeed{
		config:   config,
		handlers: handlers,
	}

	feed.leftTurnTracker = &LeftTurnTracker{
		config: config,
	}

	feed.rightTurnTracker = &RightTurnTracker{
		config: config,
	}

	feed.accelerationTracker = &AccelerationTracker{
		config: config,
	}

	feed.decelerationTracker = &DecelerationTracker{
		config: config,
	}

	feed.stopTracker = &StopTracker{
		config: config,
	}

	return feed
}

func (f *DirectionEventFeed) HandleGnssData(data *neom9n.Data) error {
	f.gnssData = data
	return nil
}

func (f *DirectionEventFeed) HandleOrientedAcceleration(acceleration *imu.Acceleration, tiltAngles *imu.TiltAngles, orientation imu.Orientation) error {
	if e := f.leftTurnTracker.track(acceleration, tiltAngles, orientation, f.gnssData); e != nil {
		if err := f.emit(e); err != nil {
			return fmt.Errorf("emitting left turn event: %w", err)
		}
	}

	if e := f.rightTurnTracker.track(acceleration, tiltAngles, orientation, f.gnssData); e != nil {
		if err := f.emit(e); err != nil {
			return fmt.Errorf("emitting right turn event: %w", err)
		}
	}
	if e := f.accelerationTracker.track(acceleration, tiltAngles, orientation, f.gnssData); e != nil {
		if err := f.emit(e); err != nil {
			return fmt.Errorf("emitting acceleration event: %w", err)
		}
	}
	if e := f.decelerationTracker.track(acceleration, tiltAngles, orientation, f.gnssData); e != nil {
		if err := f.emit(e); err != nil {
			return fmt.Errorf("emitting deceleration event: %w", err)
		}
	}
	if e := f.stopTracker.track(acceleration, tiltAngles, orientation, f.gnssData); e != nil {
		if err := f.emit(e); err != nil {
			return fmt.Errorf("emitting stop event: %w", err)
		}
	}

	return nil
}

func (f *DirectionEventFeed) emit(event data.Event) error {
	for _, handler := range f.handlers {
		err := handler(event)
		if err != nil {
			return fmt.Errorf("calling handler: %w", err)
		}
	}
	return nil
}
