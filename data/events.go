package data

import (
	"time"
)

type Event interface {
	SetTime(time.Time)
	String() string
}

type BaseEvent struct {
	time time.Time
}

func (e *BaseEvent) String() string {
	return "BaseEvent"
}

func (e *BaseEvent) SetTime(t time.Time) {
	e.time = t
}
