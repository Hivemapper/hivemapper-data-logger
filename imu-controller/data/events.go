package data

import (
	"time"

	"github.com/streamingfast/hm-imu-logger/device/iim42652"
)

type Event interface {
	setTime(time.Time)
}

type BaseEvent struct {
	time time.Time
}

func (e *BaseEvent) setTime(t time.Time) {
	e.time = t
}

type ImuAccelerationEvent struct {
	BaseEvent
	Acceleration *iim42652.Acceleration
	AvgX         *AverageFloat64
	AvgY         *AverageFloat64
	AvgZ         *AverageFloat64
}

type Direction string

const (
	Left  Direction = "left"
	Right Direction = "right"
)

type TurnEvent struct {
	BaseEvent
	Direction Direction
	Duration  time.Duration
}

type AccelerationEvent struct {
	BaseEvent
	Speed    float64
	Duration time.Duration
}

type DecelerationEvent struct {
	BaseEvent
	Speed    float64
	Duration time.Duration
}

type HeadingChangeEvent struct {
	BaseEvent
	Heading float64
}

type StopEvent struct {
	BaseEvent
	Duration time.Duration
}

type emit func(event Event)
type HandleEvent func(event Event)
type EventEmitter struct {
	eventHandler HandleEvent
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{}
}

func (e *EventEmitter) emit(event Event) {
	event.setTime(time.Now())
	e.eventHandler(event)
}

func (e *EventEmitter) Run(p *AccelerationPipeline) (err error) {
	sub := p.SubscribeAcceleration("event-emitter")
	xAvg := NewAverageFloat64("X average")
	yAvg := NewAverageFloat64("Y average")
	zAvg := NewAverageFloat64("Y average")
	totalMagnitudeAvg := NewAverageFloat64("Total magnitude average")

	leftTurnTracker := &LeftTurnTracker{
		emitFunc: e.emit,
	}

	rightTurnTracker := &RightTurnTracker{
		emitFunc: e.emit,
	}

	accelerationTracker := &AccelerationTracker{
		emitFunc: e.emit,
	}

	decelerationTracker := &DecelerationTracker{
		emitFunc: e.emit,
	}

	stopTracker := &StopTracker{
		emitFunc: e.emit,
	}

	lastUpdate := time.Time{}
	for {
		select {
		case acceleration := <-sub.IncomingAcceleration:
			if lastUpdate == (time.Time{}) {
				lastUpdate = time.Now()
				continue
			}
			//timeSinceLastUpdate := time.Since(lastUpdate)
			//speedVariation := computeSpeedVariation(timeSinceLastUpdate.Seconds(), acceleration.CamX())

			totalMagnitudeAvg.Add(computeTotalMagnitude(acceleration.CamX(), acceleration.CamY()))

			xAvg.Add(acceleration.CamX())
			yAvg.Add(acceleration.CamY())
			zAvg.Add(acceleration.CamX())

			e.emit(&ImuAccelerationEvent{
				Acceleration: acceleration,
				AvgX:         xAvg,
				AvgY:         yAvg,
				AvgZ:         zAvg,
			})

			leftTurnTracker.track(lastUpdate, acceleration, xAvg, yAvg, totalMagnitudeAvg)
			rightTurnTracker.track(lastUpdate, acceleration, xAvg, yAvg, totalMagnitudeAvg)
			accelerationTracker.track(lastUpdate, acceleration, xAvg, yAvg, totalMagnitudeAvg)
			decelerationTracker.track(lastUpdate, acceleration, xAvg, yAvg, totalMagnitudeAvg)
			stopTracker.track(acceleration, xAvg, yAvg, totalMagnitudeAvg)

			lastUpdate = time.Now()
		}
	}
}

func (e *EventEmitter) RegisterEventHandler(h HandleEvent) {
	e.eventHandler = h
}
