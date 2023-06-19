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

//func (c *OrientationCounter) String() string {
//	return fmt.Sprintf("%v", c)
//}

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
var lastKnownOrientation = OrientationUnset
var first = true

func (f *OrientedAccelerationFeed) HandleTiltCorrectedAcceleration(acceleration *Acceleration, tiltAngles *TiltAngles) error {
	//todo: stop lock for orientation when confident
	if first {
		first = false
		fmt.Println("First orientation event:", acceleration.Time)

	}
	//g += 1
	newOrientation := computeOrientation(acceleration)
	//fmt.Println("Orientation:", newOrientation, "???", f.orientationCounter.Orientation(), counter)
	if lastKnownOrientation != f.orientationCounter.Orientation() {
		lastKnownOrientation = f.orientationCounter.Orientation()
		fmt.Println("Orientation changed:", lastKnownOrientation, acceleration.Time, f.orientationCounter)
	}

	if f.orientationCounter.Orientation() != OrientationUnset {
		a := NewAcceleration(acceleration.X, acceleration.Y, acceleration.Z, acceleration.Magnitude, acceleration.Temperature, acceleration.Time)
		a = FixAccelerationOrientation(a, f.orientationCounter.Orientation())
		t := FixTiltOrientation(tiltAngles, f.orientationCounter.Orientation())

		for _, handler := range f.handlers {
			err := handler(a, t, f.orientationCounter.Orientation())
			if err != nil {
				return fmt.Errorf("calling handler: %w", err)
			}
		}
		return nil
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
	//movementThreshold := 0.1
	frontDetectionThreshold := 0.04
	backDetectionThreshold := -0.1
	rightDetectionThreshold := 0.1
	leftDetectionThreshold := -0.1

	o := OrientationUnset
	if x > frontDetectionThreshold {
		o = OrientationFront
	} else if y > rightDetectionThreshold {
		o = OrientationRight
	} else if y < leftDetectionThreshold {
		o = OrientationLeft
	} else if x < backDetectionThreshold {
		o = OrientationBack
	}
	//fmt.Println("computeOrientation", acceleration.Time, x, y, o)
	//fmt.Fprintf(os.Stderr, "%.5f,%.5f,%.5f, %.5f, %s\n", x, y, acceleration.Z, acceleration.Magnitude, o)
	return o
}
