package imu

import (
	"fmt"
)

type OrientationCounter map[Orientation]int

func (c OrientationCounter) Increment(o Orientation) {
	if _, found := c[o]; !found {
		c[o] = 1
		return
	}
	c[o] = c[o] + 1
}

func (c OrientationCounter) Orientation() Orientation {
	max := 0
	var orientation Orientation
	for o, count := range c {
		if count > max {
			max = count
			orientation = o
		}
	}
	if orientation == "" {
		return OrientationUnset
	}

	return orientation
}

type OrientedAccelerationHandler func(corrected *Acceleration, tiltAngles *TiltAngles, orientation Orientation) error

type OrientedAccelerationFeed struct {
	orientationCounter OrientationCounter
	handlers           []OrientedAccelerationHandler
}

func NewOrientedAccelerationFeed(handlers ...OrientedAccelerationHandler) *OrientedAccelerationFeed {
	return &OrientedAccelerationFeed{
		orientationCounter: make(OrientationCounter),
		handlers:           handlers,
	}
}

var g = 0
var counter = 0
var lastOrientation = OrientationUnset

func (f *OrientedAccelerationFeed) HandleTiltCorrectedAcceleration(acceleration *Acceleration, tiltAngles *TiltAngles) error {

	//todo: stop lock for orientation when confident

	//g += 1
	newOrientation := computeOrientation(acceleration)
	fmt.Println("Orientation:", newOrientation, "???", f.orientationCounter.Orientation(), counter)
	if f.orientationCounter.Orientation() != OrientationUnset {

		a := NewAcceleration(acceleration.X, acceleration.Y, acceleration.Z, acceleration.Magnitude, acceleration.Time)
		a = FixAccelerationOrientation(a, f.orientationCounter.Orientation())
		t := FixTiltOrientation(tiltAngles, f.orientationCounter.Orientation())

		for _, handler := range f.handlers {
			err := handler(a, t, f.orientationCounter.Orientation())
			if err != nil {
				return fmt.Errorf("calling handler: %w", err)
			}
		}
	}

	if newOrientation == OrientationUnset {
		lastOrientation = OrientationUnset
		counter = 0
		return nil
	}

	if newOrientation != lastOrientation && lastOrientation != OrientationUnset {
		lastOrientation = newOrientation
		counter = 0
		return nil
	}

	counter++
	if counter > 20 {
		f.orientationCounter.Increment(newOrientation)
	}

	lastOrientation = newOrientation

	return nil
}

func computeOrientation(acceleration *Acceleration) Orientation {
	x := acceleration.X
	y := acceleration.Y

	movementThreshold := 0.1
	backDetectionThreshold := -0.1
	rightDetectionThreshold := 0.1
	leftDetectionThreshold := -0.1

	if x > movementThreshold {
		return OrientationFront
	} else if y > rightDetectionThreshold {
		return OrientationRight
	} else if y < leftDetectionThreshold {
		return OrientationLeft
	} else if x < backDetectionThreshold {
		return OrientationBack
	}

	return OrientationUnset
}
