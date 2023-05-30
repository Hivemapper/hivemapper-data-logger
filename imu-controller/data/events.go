package data

import (
	"fmt"
	"math"
	"time"

	"github.com/streamingfast/hm-imu-logger/device/iim42652"
)

type Event interface {
}

type ImuAccelerationEvent struct {
	Event
	acceleration *iim42652.Acceleration
}

type Direction string

const (
	Left  Direction = "left"
	Right Direction = "right"
)

type TurnEvent struct {
	Event
	Direction Direction
	Duration  time.Duration
}

type AccelerationEvent struct {
	Event
	Speed    float64
	Duration time.Duration
}

type DecelerationEvent struct {
	Event
	Speed    float64
	Duration time.Duration
}

type HeadingChangeEvent struct {
	Event
	Heading float64
}

type StopEvent struct {
	Event
	Duration time.Duration
}

type emit func(event Event)

type EventEmitter struct {
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{}
}

func (e *EventEmitter) emit(event Event) {
	fmt.Println(event)
}

func (e *EventEmitter) Run(p *AccelerationPipeline) (err error) {
	sub := p.SubscribeAcceleration("event-emitter")
	xAvg := NewAverageFloat64("X average")
	yAvg := NewAverageFloat64("Y average")
	totalMagnitudeAvg := NewAverageFloat64("Total magnitude average")

	leftTurnTracker := &LeftTurnTracker{
		emitFunc: e.emit,
	}

	for {
		lastUpdate := time.Time{}
		select {
		case acceleration := <-sub.IncomingAcceleration:
			if lastUpdate == (time.Time{}) {
				lastUpdate = time.Now()
				continue
			}
			//timeSinceLastUpdate := time.Since(lastUpdate)
			//speedVariation := computeSpeedVariation(timeSinceLastUpdate.Seconds(), acceleration.CamX())

			totalMagnitudeAvg.Add(math.Sqrt(math.Pow(acceleration.CamX(), 2) + math.Pow(acceleration.CamY(), 2)))

			xAvg.Add(acceleration.CamX())
			yAvg.Add(acceleration.CamY())

			leftTurnTracker.track(acceleration, xAvg, yAvg, totalMagnitudeAvg)

			lastUpdate = time.Now()
		}
	}
}
