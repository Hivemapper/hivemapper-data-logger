package data

import (
	"fmt"
	"time"

	"github.com/streamingfast/hm-imu-logger/device/iim42652"
)

type Event interface {
	setTime(time.Time)
	String() string
}

type BaseEvent struct {
	time time.Time
}

func (e *BaseEvent) String() string {
	return "base"
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
	AvgMagnitude *AverageFloat64
}

func (e *ImuAccelerationEvent) String() string {
	return "imu-acceleration"
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

func (e *TurnEvent) String() string {
	return fmt.Sprintf("turn %s in %s", e.Direction, e.Duration)
}

type AccelerationEvent struct {
	BaseEvent
	Speed    float64
	Duration time.Duration
}

func (e *AccelerationEvent) String() string {
	return fmt.Sprintf("acceleration %f km/h in %s", e.Speed, e.Duration)
}

type DecelerationEvent struct {
	BaseEvent
	Speed    float64
	Duration time.Duration
}

func (e *DecelerationEvent) String() string {
	return fmt.Sprintf("deceleration %f km/h in %s", e.Speed, e.Duration)
}

type HeadingChangeEvent struct {
	BaseEvent
	Heading float64
}

func (e *HeadingChangeEvent) String() string {
	return fmt.Sprintf("heading change %f", e.Heading)
}

type StopDetectEvent struct {
	BaseEvent
}

func (e *StopDetectEvent) String() string {
	return "stop detect"
}

type StopEndEvent struct {
	BaseEvent
	Duration time.Duration
}

func (e *StopEndEvent) String() string {
	return fmt.Sprintf("stop for %s", e.Duration)
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
	if e.eventHandler == nil {
		return
	}
	event.setTime(time.Now())
	e.eventHandler(event)
}

func (e *EventEmitter) Run(p *AccelerationPipeline) (err error) {
	sub := p.SubscribeAcceleration("event-emitter")
	xAvg := NewAverageFloat64("X average")
	yAvg := NewAverageFloat64("Y average")
	zAvg := NewAverageFloat64("Y average")
	magnitudeAvg := NewAverageFloat64("Total magnitude average")

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

			magnitudeAvg.Add(computeTotalMagnitude(acceleration.CamX(), acceleration.CamY()))

			xAvg.Add(acceleration.CamX())
			yAvg.Add(acceleration.CamY())
			zAvg.Add(acceleration.CamX())

			e.emit(&ImuAccelerationEvent{
				Acceleration: acceleration,
				AvgX:         xAvg,
				AvgY:         yAvg,
				AvgZ:         zAvg,
				AvgMagnitude: magnitudeAvg,
			})

			leftTurnTracker.track(lastUpdate, acceleration, xAvg, yAvg, magnitudeAvg)
			rightTurnTracker.track(lastUpdate, acceleration, xAvg, yAvg, magnitudeAvg)
			accelerationTracker.track(lastUpdate, acceleration, xAvg, yAvg, magnitudeAvg)
			decelerationTracker.track(lastUpdate, acceleration, xAvg, yAvg, magnitudeAvg)
			stopTracker.track(acceleration, xAvg, yAvg, magnitudeAvg)

			lastUpdate = time.Now()
		}
	}
}

func (e *EventEmitter) RegisterEventHandler(h HandleEvent) {
	e.eventHandler = h
}
