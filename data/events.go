package data

import (
	"fmt"
	"math"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
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
	Time     time.Time    `json:"time"`
	Name     string       `json:"name"`
	Category string       `json:"category"`
	GnssData *neom9n.Data `json:"gnss_data"`
}

func NewBaseEvent(name string, category string, time time.Time, gnssData *neom9n.Data) *BaseEvent {
	return &BaseEvent{
		Name:     name,
		Category: category,
		Time:     time,
		GnssData: gnssData,
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
	return e.GnssData
}

func (e *BaseEvent) GetCategory() string {
	return e.Category
}
