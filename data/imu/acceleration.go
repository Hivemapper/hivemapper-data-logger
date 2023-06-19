package imu

import (
	"fmt"
	"math"
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

func NewTiltAngles(x, y, z float64) *TiltAngles {
	return &TiltAngles{
		X: x,
		Y: y,
		Z: z,
	}
}

type Acceleration struct {
	X         float64
	Y         float64
	Z         float64
	Magnitude float64
	Time      time.Time
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

func FixAccelerationOrientation(acceleration *Acceleration, orientation Orientation) *Acceleration {
	return NewAcceleration(
		fixX(acceleration, orientation),
		fixY(acceleration, orientation),
		acceleration.Z,
		acceleration.Magnitude,
		acceleration.Time)
}

func FixTiltOrientation(tilt *TiltAngles, orientation Orientation) *TiltAngles {
	return NewTiltAngles(
		fixXAngle(tilt, orientation),
		fixYAngle(tilt, orientation),
		tilt.Z,
	)
}

func fixX(acceleration *Acceleration, orientation Orientation) float64 {
	switch orientation {
	case OrientationFront:
		return acceleration.X
	case OrientationRight:
		return invert(acceleration.Y)
	case OrientationLeft:
		return acceleration.Y
	case OrientationBack:
		return invert(acceleration.X)
	default:
		panic(fmt.Sprintf("invalid orientation %q", orientation))
	}
}

func fixXAngle(tilt *TiltAngles, orientation Orientation) float64 {
	switch orientation {
	case OrientationFront:
		return tilt.X
	case OrientationRight:
		return invert(tilt.Y)
	case OrientationLeft:
		return tilt.Y
	case OrientationBack:
		return invert(tilt.X)
	default:
		panic("invalid orientation")
	}
}

func fixY(acceleration *Acceleration, orientation Orientation) float64 {
	switch orientation {
	case OrientationFront:
		return acceleration.Y
	case OrientationRight:
		return acceleration.X
	case OrientationLeft:
		return invert(acceleration.X)
	case OrientationBack:
		return invert(acceleration.Y)
	default:
		panic("invalid orientation")
	}
}

func fixYAngle(tilt *TiltAngles, orientation Orientation) float64 {
	switch orientation {
	case OrientationFront:
		return tilt.Y
	case OrientationRight:
		return tilt.X
	case OrientationLeft:
		return invert(tilt.X)
	case OrientationBack:
		return invert(tilt.Y)
	case OrientationUnset:
		return math.MaxFloat64
	default:
		panic("invalid orientation")
	}
}

func invert(val float64) float64 {
	return -val
}
