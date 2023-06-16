package data

import (
	"fmt"
	"math"
	"time"

	"github.com/streamingfast/gnss-controller/device/neom9n"
)

func round(v float64) float64 {
	return math.Round(v*100) / 100
}

type Event interface {
	SetTime(time.Time)
	GetTime() time.Time
	String() string
	GetName() string
	GetCategory() string
	GetGnssData() *neom9n.Data
}

type BaseEvent struct {
	Time     time.Time `json:"time"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
	gnssData *neom9n.Data
}

func NewBaseEvent(name string, category string, time time.Time, gnssData *neom9n.Data) *BaseEvent {
	return &BaseEvent{
		Name:     name,
		Category: category,
		Time:     time,
		gnssData: gnssData,
	}
}

func (e *BaseEvent) String() string {
	return fmt.Sprintf("BaseEvent: %s @ %s", e.Name, e.Category)
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

func (e *BaseEvent) GetGnssData() *neom9n.Data {
	return e.gnssData
}

func (e *BaseEvent) GetCategory() string {
	return e.Category
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
