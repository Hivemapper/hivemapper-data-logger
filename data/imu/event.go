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
	AvgX         *data.AverageFloat64   `json:"avg_x"`
	AvgY         *data.AverageFloat64   `json:"avg_y"`
	AvgZ         *data.AverageFloat64   `json:"avg_z"`
	AvgMagnitude *data.AverageFloat64   `json:"avg_magnitude"`
}

func (e *ImuAccelerationEvent) String() string {
	return "ImuAccelerationEvent"
}

func NewImuAccelerationEvent(acc *iim42652.Acceleration, xAvg *data.AverageFloat64, yAvg *data.AverageFloat64, zAvg *data.AverageFloat64, avgMag *data.AverageFloat64) *ImuAccelerationEvent {
	return &ImuAccelerationEvent{
		BaseEvent:    data.NewBaseEvent("IMU_ACCELERATION_EVENT", data.NewGForcePosition(xAvg.Average, yAvg.Average, zAvg.Average)),
		Acceleration: acc,
		AvgX:         xAvg,
		AvgY:         yAvg,
		AvgZ:         zAvg,
		AvgMagnitude: avgMag,
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
