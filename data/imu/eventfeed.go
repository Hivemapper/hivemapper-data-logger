package imu

import (
	"fmt"
	"time"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type emit func(event data.Event)

type subscriptions map[string]*data.Subscription

type EventFeed struct {
	imu           *iim42652.IIM42652
	subscriptions subscriptions
	config        *Config
}

func NewEventFeed(imu *iim42652.IIM42652, config *Config) *EventFeed {
	return &EventFeed{
		config:        config,
		imu:           imu,
		subscriptions: make(subscriptions),
	}
}

func (p *EventFeed) Run() error {
	fmt.Println("Running pipeline")
	err := p.run()
	if err != nil {
		return fmt.Errorf("running pipeline: %w", err)
	}
	return nil
}

func (p *EventFeed) Subscribe(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	p.subscriptions[name] = sub
	return sub
}

func (p *EventFeed) emit(event data.Event) {
	event.SetTime(time.Now())
	for _, subscription := range p.subscriptions {
		subscription.IncomingEvents <- event
	}
}

func (p *EventFeed) run() error {
	xAvg := data.NewAverageFloat64("X average")
	yAvg := data.NewAverageFloat64("Y average")
	zAvg := data.NewAverageFloat64("Y average")
	magnitudeAvg := data.NewAverageFloat64("Total magnitude average")

	leftTurnTracker := &LeftTurnTracker{
		config:   p.config,
		emitFunc: p.emit,
	}

	rightTurnTracker := &RightTurnTracker{
		config:   p.config,
		emitFunc: p.emit,
	}

	accelerationTracker := &AccelerationTracker{
		config:   p.config,
		emitFunc: p.emit,
	}

	decelerationTracker := &DecelerationTracker{
		config:   p.config,
		emitFunc: p.emit,
	}

	stopTracker := &StopTracker{
		config:   p.config,
		emitFunc: p.emit,
	}

	lastUpdate := time.Time{}

	for {
		acceleration, err := p.imu.GetAcceleration()
		if err != nil {
			panic(fmt.Errorf("getting acceleration: %w", err))
		}

		magnitudeAvg.Add(computeTotalMagnitude(acceleration.CamX(), acceleration.CamY()))

		xAvg.Add(acceleration.CamX())
		yAvg.Add(acceleration.CamY())
		zAvg.Add(acceleration.CamX())

		p.emit(&ImuAccelerationEvent{
			Acceleration: acceleration,
			AvgX:         xAvg,
			AvgY:         yAvg,
			AvgZ:         zAvg,
			AvgMagnitude: magnitudeAvg,
		})

		x := xAvg.Average
		y := yAvg.Average
		z := zAvg.Average

		leftTurnTracker.trackAcceleration(lastUpdate, x, y, z)
		rightTurnTracker.trackAcceleration(lastUpdate, x, y, z)
		accelerationTracker.trackAcceleration(lastUpdate, x, y, z)
		decelerationTracker.trackAcceleration(lastUpdate, x, y, z)
		stopTracker.trackAcceleration(lastUpdate, x, y, y)

		lastUpdate = time.Now()

		time.Sleep(10 * time.Millisecond)
	}
}
