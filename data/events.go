package data

import (
	"time"
)

type Event interface {
	SetTime(time.Time)
	GetTime() time.Time
	String() string
}

type BaseEvent struct {
	Time time.Time
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
