package data

import (
	"fmt"
	"math"
	"time"
)

func round(v float64) float64 {
	return math.Round(v*100) / 100
}

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

func (g *GForcePosition) String() string {
	return fmt.Sprintf("xAvg: %.10f yAvg: %.10f zAvg: %.10f", round(g.XAvg), round(g.YAvg), round(g.ZAvg))
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
