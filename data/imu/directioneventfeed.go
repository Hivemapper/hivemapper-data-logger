package imu

import (
	"fmt"
	"time"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type emit func(event data.Event)

type DirectionEventFeed struct {
	imu                 *iim42652.IIM42652
	subscriptions       data.Subscriptions
	config              *Config
	leftTurnTracker     *LeftTurnTracker
	rightTurnTracker    *RightTurnTracker
	accelerationTracker *AccelerationTracker
	decelerationTracker *DecelerationTracker
	stopTracker         *StopTracker
	lastUpdate          time.Time
}

func NewDirectionEventFeed(config *Config) *DirectionEventFeed {
	feed := &DirectionEventFeed{
		config:        config,
		subscriptions: make(data.Subscriptions),
	}
	emit := feed.emit

	feed.leftTurnTracker = &LeftTurnTracker{
		config:   config,
		emitFunc: emit,
	}

	feed.rightTurnTracker = &RightTurnTracker{
		config:   config,
		emitFunc: emit,
	}

	feed.accelerationTracker = &AccelerationTracker{
		config:   config,
		emitFunc: emit,
	}

	feed.decelerationTracker = &DecelerationTracker{
		config:   config,
		emitFunc: emit,
	}

	feed.stopTracker = &StopTracker{
		config:   config,
		emitFunc: emit,
	}

	return feed
}

func (f *DirectionEventFeed) Subscribe(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	f.subscriptions[name] = sub
	return sub
}

func (f *DirectionEventFeed) Start(sub *data.Subscription) {
	fmt.Println("Running direction event feed")
	now := time.Now()
	f.lastUpdate = now

	go func() {
		for {
			select {
			case event := <-sub.IncomingEvents:
				if len(f.subscriptions) == 0 {
					continue
				}
				e := event.(*CorrectedAccelerationEvent)
				err := f.handleEvent(e)
				f.lastUpdate = e.GetTime()

				if err != nil {
					panic(fmt.Errorf("handling event %s: %w", e.GetName(), err))
				}
			}
		}
	}()
}

func (f *DirectionEventFeed) emit(event data.Event) {
	event.SetTime(time.Now())
	for _, subscription := range f.subscriptions {
		subscription.IncomingEvents <- event
	}
}

func (f *DirectionEventFeed) handleEvent(e *CorrectedAccelerationEvent) error {
	x := e.X
	y := e.Y

	f.leftTurnTracker.track(f.lastUpdate, x, y)
	f.rightTurnTracker.track(f.lastUpdate, x, y)
	f.accelerationTracker.track(f.lastUpdate, x, y)
	f.decelerationTracker.track(f.lastUpdate, x, y)
	f.stopTracker.track(f.lastUpdate, x, y)

	return nil
}
