package imu

import (
	"fmt"
	"time"
)

type Orientation string

const (
	OrientationUnset Orientation = ""
	OrientationFront Orientation = "OrientationFront"
	OrientationRight Orientation = "OrientationRight"
	OrientationLeft  Orientation = "OrientationLeft"
	OrientationBack  Orientation = "OrientationBack"
)

type TiltAngles struct {
	X float64
	Y float64
	Z float64
}

type Acceleration struct {
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	Z         float64   `json:"z"`
	Magnitude float64   `json:"magnitude"`
	Time      time.Time `json:"time"`
}

func NewAcceleration(x, y, z, m float64, time time.Time) *Acceleration {
	return &Acceleration{
		X:         x,
		Y:         y,
		Z:         z,
		Magnitude: m,
		Time:      time,
	}
}

func (a *Acceleration) String() string {
	return fmt.Sprintf("Acceleration{x=%f, y=%f, z=%f, magnitude=%f, time=%s}", a.X, a.Y, a.Z, a.Magnitude, a.Time)
}

// func invert(val float64) float64 {
// 	return -val
// }
