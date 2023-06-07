package imu

import (
	"fmt"
	"time"

	"github.com/rosshemsley/kalman"
	"github.com/rosshemsley/kalman/models"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type emit func(event data.Event)

type DirectionEventFeed struct {
	imu                 *iim42652.IIM42652
	subscriptions       data.Subscriptions
	config              *Config
	xModel              *models.SimpleModel
	xFilter             *kalman.KalmanFilter
	yModel              *models.SimpleModel
	yFilter             *kalman.KalmanFilter
	zModel              *models.SimpleModel
	zFilter             *kalman.KalmanFilter
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

func (f *DirectionEventFeed) Run(feed *CorrectedAccelerationFeed) error {
	sub := feed.Subscribe("corrected")
	now := time.Now()
	f.lastUpdate = now
	f.xModel = models.NewSimpleModel(now, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	f.xFilter = kalman.NewKalmanFilter(f.xModel)

	f.yModel = models.NewSimpleModel(now, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	f.yFilter = kalman.NewKalmanFilter(f.yModel)

	f.zModel = models.NewSimpleModel(now, 1.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	f.zFilter = kalman.NewKalmanFilter(f.zModel)

	for {
		select {
		case event := <-sub.IncomingEvents:
			if len(f.subscriptions) == 0 {
				continue
			}
			e := event.(*CorrectedAccelerationEvent)
			err := f.handleEvent(e)
			f.lastUpdate = time.Now()

			return fmt.Errorf("handling event: %w", err)
		}
	}
}

func (f *DirectionEventFeed) emit(event data.Event) {
	event.SetTime(time.Now())
	for _, subscription := range f.subscriptions {
		subscription.IncomingEvents <- event
	}
}

func (f *DirectionEventFeed) handleEvent(e *CorrectedAccelerationEvent) error {

	//x := e.X
	//y := e.Y
	//
	//now := time.Now()
	//err := f.xFilter.Update(now, f.xModel.NewMeasurement(x))
	//if err != nil {
	//	return fmt.Errorf("updating kalman x filter: %w", err)
	//}
	//
	//err = f.yFilter.Update(now, f.yModel.NewMeasurement(y))
	//if err != nil {
	//	return fmt.Errorf("updating kalman y filter: %w", err)
	//}
	//
	//err = f.zFilter.Update(now, f.zModel.NewMeasurement(z))
	//if err != nil {
	//	return fmt.Errorf("updating kalman z filter: %w", err)
	//}

	//f.leftTurnTracker.trackAcceleration(f.lastUpdate, x, y, z)
	//f.rightTurnTracker.trackAcceleration(f.lastUpdate, x, y, z)
	//f.accelerationTracker.trackAcceleration(f.lastUpdate, x, y, z)
	//f.decelerationTracker.trackAcceleration(f.lastUpdate, x, y, z)
	//f.stopTracker.trackAcceleration(f.lastUpdate, x, y, z)
	return nil
}
