package data

import (
	"fmt"
	"math"
	"time"
)

func round(v float64) float64 {
	return math.Round(v*100) / 100
}

type Event interface {
	SetTime(time.Time)
	GetTime() time.Time
	String() string
	GetName() string
}

func NewBaseEvent(name string) *BaseEvent {
	return &BaseEvent{
		Name: name,
		Time: time.Now(),
	}
}

type BaseEvent struct {
	Time time.Time `json:"time"`
	Name string    `json:"name"`
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

type Angles struct {
	xAngle float64
	yAngle float64
	zAngle float64
}

func NewAngles(x, y, z float64) *Angles {
	return &Angles{
		xAngle: x,
		yAngle: y,
		zAngle: z,
	}
}

func (a *Angles) String() string {
	return fmt.Sprintf("xAngle: %.10f yAngle: %.10f zAngle: %.10f", round(a.xAngle), round(a.yAngle), round(a.zAngle))
}
