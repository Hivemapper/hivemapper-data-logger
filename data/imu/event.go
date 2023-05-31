package imu

import (
	"fmt"
	"time"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type ImuAccelerationEvent struct {
	data.BaseEvent
	Acceleration *iim42652.Acceleration
	AvgX         *data.AverageFloat64
	AvgY         *data.AverageFloat64
	AvgZ         *data.AverageFloat64
	AvgMagnitude *data.AverageFloat64
}

func (e *ImuAccelerationEvent) String() string {
	return "ImuAccelerationEvent"
}

type Direction string

const (
	Left  Direction = "left"
	Right Direction = "right"
)

type TurnEvent struct {
	data.BaseEvent
	Direction Direction
	Duration  time.Duration
}

func (e *TurnEvent) String() string {
	return fmt.Sprintf("%s turn for %s", e.Direction, e.Duration)
}

type AccelerationDetectedEvent struct {
	data.BaseEvent
}

func (e *AccelerationDetectedEvent) String() string {
	return "Acceleration Detected"
}

type AccelerationEvent struct {
	data.BaseEvent
	Speed    float64
	Duration time.Duration
}

func (e *AccelerationEvent) String() string {
	return fmt.Sprintf("AccelerationEvent of %f km/h for %s", e.Speed, e.Duration)
}

type DecelerationEvent struct {
	data.BaseEvent
	Speed    float64
	Duration time.Duration
}

func (e *DecelerationEvent) String() string {
	return fmt.Sprintf("DecelerationEvent => %f km/h in %s", e.Speed, e.Duration)
}

type HeadingChangeEvent struct {
	data.BaseEvent
	Heading float64
}

func (e *HeadingChangeEvent) String() string {
	return fmt.Sprintf("Heading Change %f", e.Heading)
}

type StopDetectedEvent struct {
	data.BaseEvent
}

func (e *StopDetectedEvent) String() string {
	return "Stop Detected"
}

type StopEndEvent struct {
	data.BaseEvent
	Duration time.Duration
}

func (e *StopEndEvent) String() string {
	return fmt.Sprintf("Stop End for %s", e.Duration)
}
