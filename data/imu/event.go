package imu

import (
	"fmt"
	"time"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type ImuAccelerationEvent struct {
	*data.BaseEvent
	Acceleration *iim42652.Acceleration `json:"acceleration"`
	X            float64                `json:"x"`
	Y            float64                `json:"y"`
	Z            float64                `json:"z"`
	AvgMagnitude float64                `json:"magnitude"`
}

func (e *ImuAccelerationEvent) String() string {
	return "ImuAccelerationEvent"
}

func NewImuAccelerationEvent(acc *iim42652.Acceleration, x float64, y float64, z float64, magnitude float64) *ImuAccelerationEvent {
	return &ImuAccelerationEvent{
		BaseEvent:    data.NewBaseEvent("IMU_ACCELERATION_EVENT", data.NewGForcePosition(x, y, z)),
		Acceleration: acc,
		X:            x,
		Y:            y,
		Z:            z,
		AvgMagnitude: magnitude,
	}
}

type RightTurnEventDetected struct {
	*data.BaseEvent
}

func NewRightTurnEventDetected(gForcePosition *data.GForcePosition) *RightTurnEventDetected {
	return &RightTurnEventDetected{
		BaseEvent: data.NewBaseEvent("RIGHT_TURN_DETECTED_EVENT", gForcePosition),
	}
}

func (e *RightTurnEventDetected) String() string {
	return "Right Turn Event Detected"
}

type RightTurnEvent struct {
	*data.BaseEvent
	Duration time.Duration `json:"duration"`
}

func NewRightTurnEvent(duration time.Duration, gForcePosition *data.GForcePosition) *RightTurnEvent {
	return &RightTurnEvent{
		BaseEvent: data.NewBaseEvent("RIGHT_TURN_EVENT", gForcePosition),
		Duration:  duration,
	}
}
func (e *RightTurnEvent) String() string {
	return fmt.Sprintf("Right Turn Event for %s", e.Duration)
}

type LeftTurnEventDetected struct {
	*data.BaseEvent
}

func NewLeftTurnEventDetected(gForcePosition *data.GForcePosition) *LeftTurnEventDetected {
	return &LeftTurnEventDetected{
		BaseEvent: data.NewBaseEvent("LEFT_TURN_DETECTED_EVENT", gForcePosition),
	}
}

func (e *LeftTurnEventDetected) String() string {
	return "Left Turn Event Detected"
}

type LeftTurnEvent struct {
	*data.BaseEvent
	Duration time.Duration `json:"duration"`
}

func NewLeftTurnEvent(duration time.Duration, gForcePosition *data.GForcePosition) *LeftTurnEvent {
	return &LeftTurnEvent{
		BaseEvent: data.NewBaseEvent("LEFT_TURN_EVENT", gForcePosition),
		Duration:  duration,
	}
}
func (e *LeftTurnEvent) String() string {
	return fmt.Sprintf("Left Turn Event for %s", e.Duration)
}

type AccelerationDetectedEvent struct {
	*data.BaseEvent
}

func NewAccelerationDetectedEvent(gForcePosition *data.GForcePosition) *AccelerationDetectedEvent {
	return &AccelerationDetectedEvent{
		BaseEvent: data.NewBaseEvent("ACCELERATION_DETECTED_EVENT", gForcePosition),
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

func NewAccelerationEvent(speed float64, duration time.Duration, gForcePosition *data.GForcePosition) *AccelerationEvent {
	return &AccelerationEvent{
		BaseEvent: data.NewBaseEvent("ACCELERATION_EVENT", gForcePosition),
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

func NewDecelerationDetectedEvent(gForcePosition *data.GForcePosition) *DecelerationDetectedEvent {
	return &DecelerationDetectedEvent{
		BaseEvent: data.NewBaseEvent("DECELERATION_DETECTED_EVENT", gForcePosition),
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

func NewDecelerationEvent(speed float64, duration time.Duration, gForcePosition *data.GForcePosition) *DecelerationEvent {
	return &DecelerationEvent{
		BaseEvent: data.NewBaseEvent("DECELERATION_EVENT", gForcePosition),
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

func NewHeadingChangeEvent(gForcePosition *data.GForcePosition) *HeadingChangeEvent {
	return &HeadingChangeEvent{
		BaseEvent: data.NewBaseEvent("HEADING_CHANGE_EVENT", gForcePosition),
	}
}

func (e *HeadingChangeEvent) String() string {
	return fmt.Sprintf("Heading Change %f", e.Heading)
}

type StopDetectedEvent struct {
	*data.BaseEvent
}

func NewStopDetectedEvent(gForcePosition *data.GForcePosition) *StopDetectedEvent {
	return &StopDetectedEvent{
		BaseEvent: data.NewBaseEvent("STOP_DETECTED_EVENT", gForcePosition),
	}
}

func (e *StopDetectedEvent) String() string {
	return "Stop Detected"
}

type StopEndEvent struct {
	*data.BaseEvent
	Duration time.Duration `json:"duration"`
}

func NewStopEndEvent(duration time.Duration, gForcePosition *data.GForcePosition) *StopEndEvent {
	return &StopEndEvent{
		BaseEvent: data.NewBaseEvent("STOP_END_EVENT", gForcePosition),
		Duration:  duration,
	}
}

func (e *StopEndEvent) String() string {
	return fmt.Sprintf("Stop End for %s", e.Duration)
}

type CorrectedEvent struct {
	*data.BaseEvent
	correctedData *data.GForcePosition
	angles        *data.Angles
}

func NewCorrectedEvent(duration time.Duration, gForcePosition *data.GForcePosition, correctedGForcePosition *data.GForcePosition, angles *data.Angles) *CorrectedEvent {
	return &CorrectedEvent{
		BaseEvent:     data.NewBaseEvent("CORRECTED_EVENT", gForcePosition),
		correctedData: correctedGForcePosition,
		angles:        angles,
	}
}

func (e *CorrectedEvent) String() string {
	return fmt.Sprintf("[Original GForce position] %s [Corrected GForce position] %s [Angles] %s", e.BaseEvent.GForcePosition.String(), e.correctedData.String(), e.angles.String())
}
