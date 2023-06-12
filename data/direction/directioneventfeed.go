package direction

import (
	"fmt"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"time"

	"github.com/streamingfast/hivemapper-data-logger/data"
)

type emit func(event data.Event)

type DirectionEventFeed struct {
	subscriptions       data.Subscriptions
	config              *imu.Config
	leftTurnTracker     *LeftTurnTracker
	rightTurnTracker    *RightTurnTracker
	accelerationTracker *AccelerationTracker
	decelerationTracker *DecelerationTracker
	stopTracker         *StopTracker
}

func NewDirectionEventFeed(config *imu.Config) *DirectionEventFeed {
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

func (f *DirectionEventFeed) Start(imuCorrectedAccelerationSub *data.Subscription, gnssSub *data.Subscription) {
	fmt.Println("Running direction event feed")

	go func() {
		var imuEvent *imu.TiltCorrectedAccelerationEvent
		var gnssEvent *gnss.GnssEvent
		for {

			select {
			case event := <-imuCorrectedAccelerationSub.IncomingEvents:
				if len(f.subscriptions) == 0 {
					continue
				}
				imuEvent = event.(*imu.TiltCorrectedAccelerationEvent)

			case event := <-gnssSub.IncomingEvents:
				if len(f.subscriptions) == 0 {
					continue
				}
				gnssEvent = event.(*gnss.GnssEvent)
			}

			if imuEvent != nil && gnssEvent != nil {
				err := f.handleEvent(imuEvent, gnssEvent)
				if err != nil {
					panic(fmt.Errorf("handling event: %w", err))
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

func (f *DirectionEventFeed) handleEvent(eventImu *imu.TiltCorrectedAccelerationEvent, eventGnss *gnss.GnssEvent) error {
	f.leftTurnTracker.track(eventImu, eventGnss)
	f.rightTurnTracker.track(eventImu, eventGnss)
	f.accelerationTracker.track(eventImu, eventGnss)
	f.decelerationTracker.track(eventImu, eventGnss)
	f.stopTracker.track(eventImu, eventGnss)

	return nil
}
