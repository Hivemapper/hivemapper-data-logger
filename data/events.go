package data

import (
	"time"
)

type GForcePosition struct {
	XAvg float64 `json:"x_avg"`
	YAvg float64 `json:"y_avg"`
	ZAvg float64 `json:"z_avg"`
}

func NewGForcePosition(x, y, z float64) *GForcePosition {
	return &GForcePosition{
		XAvg: x,
		YAvg: y,
		ZAvg: z,
	}
}

type Event interface {
	SetTime(time.Time)
	GetTime() time.Time
	String() string
	GetName() string
}

func NewBaseEvent(name string, gForcePosition *GForcePosition) *BaseEvent {
	return &BaseEvent{
		Name:           name,
		Time:           time.Now(),
		GForcePosition: gForcePosition,
	}
}

type BaseEvent struct {
	Time           time.Time       `json:"time"`
	Name           string          `json:"name"`
	GForcePosition *GForcePosition `json:"g_force_position"`
}

func (e *BaseEvent) String() string {
	return "BaseEvent"
}

func (e *BaseEvent) SetTime(t time.Time) {
	e.Time = t
}

func (e *BaseEvent) GetTime() time.Time {
	return e.Time
}

func (e *BaseEvent) GetName() string {
	return e.Name
}

func (e *BaseEvent) GetGForcePosition() *GForcePosition {
	return e.GForcePosition
}
