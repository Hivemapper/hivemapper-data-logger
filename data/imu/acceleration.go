package imu

import "math"

type Orientation string

const (
	OrientationUnset Orientation = "Unset"
	OrientationFront Orientation = "OrientationFront"
	OrientationRight Orientation = "OrientationRight"
	OrientationLeft  Orientation = "OrientationLeft"
	OrientationBack  Orientation = "OrientationBack"
)

type Acceleration struct {
	X         float64
	Y         float64
	Z         float64
	Magnitude float64
}

func NewAcceleration(x, y, z, m float64) *Acceleration {
	return &Acceleration{
		X:         x,
		Y:         y,
		Z:         z,
		Magnitude: m,
	}
}

type OrientedAcceleration struct {
	*Acceleration
	Orientation Orientation
}

func FixAccelerationOrientation(acceleration *Acceleration, orientation Orientation) *Acceleration {
	return NewAcceleration(
		fixX(acceleration, orientation),
		fixY(acceleration, orientation),
		acceleration.Z,
		acceleration.Magnitude)
}

func NewOrientedAcceleration(acceleration *Acceleration, orientation Orientation) *OrientedAcceleration {
	a := NewAcceleration(
		acceleration.X,
		acceleration.Y,
		acceleration.Z,
		acceleration.Magnitude)
	return &OrientedAcceleration{
		Acceleration: a,
		Orientation:  orientation,
	}
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
	case OrientationUnset:
		return math.MaxFloat64
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
	case OrientationUnset:
		return math.MaxFloat64
	default:
		panic("invalid orientation")
	}
}

func invert(val float64) float64 {
	return -val
}

type TiltCorrectedAcceleration struct {
	*OrientedAcceleration
	XAngle float64
	YAngle float64
	ZAngle float64
}

func NewTiltCorrectedAcceleration(orientedAcceleration *OrientedAcceleration, xAngle, yAngle, zAngle float64) *TiltCorrectedAcceleration {
	return &TiltCorrectedAcceleration{
		OrientedAcceleration: orientedAcceleration,
		XAngle:               xAngle,
		YAngle:               yAngle,
		ZAngle:               zAngle,
	}
}
