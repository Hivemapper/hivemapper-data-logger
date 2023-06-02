package data

import (
	"time"
)

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
