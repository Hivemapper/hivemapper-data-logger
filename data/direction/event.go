package direction

import (
	"fmt"
	"time"

	"github.com/streamingfast/hivemapper-data-logger/data"
)

type RightTurnEventDetected struct {
	*data.BaseEvent
}

func NewRightTurnEventDetected() *RightTurnEventDetected {
	return &RightTurnEventDetected{
		BaseEvent: data.NewBaseEvent("RIGHT_TURN_DETECTED_EVENT", "DIRECTION_CHANGE"),
	}
}

func (e *RightTurnEventDetected) String() string {
	return "OrientationRight Turn Event Detected"
}

type RightTurnEvent struct {
	*data.BaseEvent
	Duration time.Duration `json:"duration"`
}

func NewRightTurnEvent(duration time.Duration) *RightTurnEvent {
	return &RightTurnEvent{
		BaseEvent: data.NewBaseEvent("RIGHT_TURN_EVENT", "DIRECTION_CHANGE"),
		Duration:  duration,
	}
}
func (e *RightTurnEvent) String() string {
	return fmt.Sprintf("OrientationRight Turn Event for %s", e.Duration)
}

type LeftTurnEventDetected struct {
	*data.BaseEvent
}

func NewLeftTurnEventDetected() *LeftTurnEventDetected {
	return &LeftTurnEventDetected{
		BaseEvent: data.NewBaseEvent("LEFT_TURN_DETECTED_EVENT", "DIRECTION_CHANGE"),
	}
}

func (e *LeftTurnEventDetected) String() string {
	return "OrientationLeft Turn Event Detected"
}

type LeftTurnEvent struct {
	*data.BaseEvent
	Duration time.Duration `json:"duration"`
}

func NewLeftTurnEvent(duration time.Duration) *LeftTurnEvent {
	return &LeftTurnEvent{
		BaseEvent: data.NewBaseEvent("LEFT_TURN_EVENT", "DIRECTION_CHANGE"),
		Duration:  duration,
	}
}
func (e *LeftTurnEvent) String() string {
	return fmt.Sprintf("OrientationLeft Turn Event for %s", e.Duration)
}

type AccelerationDetectedEvent struct {
	*data.BaseEvent
}

func NewAccelerationDetectedEvent() *AccelerationDetectedEvent {
	return &AccelerationDetectedEvent{
		BaseEvent: data.NewBaseEvent("ACCELERATION_DETECTED_EVENT", "DIRECTION_CHANGE"),
	}
}
func (e *AccelerationDetectedEvent) String() string {
	return "Acceleration Detected"
}

type AccelerationEvent struct {
	*data.BaseEvent
	Speed    float64       `json:"speed"`
	Duration time.Duration `json:"duration"`
}

func NewAccelerationEvent(speed float64, duration time.Duration) *AccelerationEvent {
	return &AccelerationEvent{
		BaseEvent: data.NewBaseEvent("ACCELERATION_EVENT", "DIRECTION_CHANGE"),
		Speed:     speed,
		Duration:  duration,
	}
}

func (e *AccelerationEvent) String() string {
	return fmt.Sprintf("AccelerationEvent of %f km/h for %s", e.Speed, e.Duration)
}

type DecelerationDetectedEvent struct {
	*data.BaseEvent
}

func NewDecelerationDetectedEvent() *DecelerationDetectedEvent {
	return &DecelerationDetectedEvent{
		BaseEvent: data.NewBaseEvent("DECELERATION_DETECTED_EVENT", "DIRECTION_CHANGE"),
	}
}

func (e *DecelerationDetectedEvent) String() string {
	return "Deceleration Detected"
}

type DecelerationEvent struct {
	*data.BaseEvent
	Speed    float64       `json:"speed"`
	Duration time.Duration `json:"duration"`
}

func NewDecelerationEvent(speed float64, duration time.Duration) *DecelerationEvent {
	return &DecelerationEvent{
		BaseEvent: data.NewBaseEvent("DECELERATION_EVENT", "DIRECTION_CHANGE"),
		Speed:     speed,
		Duration:  duration,
	}
}

func (e *DecelerationEvent) String() string {
	return fmt.Sprintf("DecelerationEvent => %f km/h in %s", e.Speed, e.Duration)
}

type HeadingChangeEvent struct {
	*data.BaseEvent
	Heading float64 `json:"heading"`
}

func NewHeadingChangeEvent() *HeadingChangeEvent {
	return &HeadingChangeEvent{
		BaseEvent: data.NewBaseEvent("HEADING_CHANGE_EVENT", "DIRECTION_CHANGE"),
	}
}

func (e *HeadingChangeEvent) String() string {
	return fmt.Sprintf("Heading Change %f", e.Heading)
}

type StopDetectedEvent struct {
	*data.BaseEvent
}

func NewStopDetectedEvent() *StopDetectedEvent {
	return &StopDetectedEvent{
		BaseEvent: data.NewBaseEvent("STOP_DETECTED_EVENT", "DIRECTION_CHANGE"),
	}
}

func (e *StopDetectedEvent) String() string {
	return "Stop Detected"
}

type StopEndEvent struct {
	*data.BaseEvent
	Duration time.Duration `json:"duration"`
}

func NewStopEndEvent(duration time.Duration) *StopEndEvent {
	return &StopEndEvent{
		BaseEvent: data.NewBaseEvent("STOP_END_EVENT", "DIRECTION_CHANGE"),
		Duration:  duration,
	}
}

func (e *StopEndEvent) String() string {
	return fmt.Sprintf("Stop End for %s", e.Duration)
}
