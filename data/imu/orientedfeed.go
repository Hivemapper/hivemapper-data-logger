package imu

import (
	"fmt"

	"github.com/streamingfast/hivemapper-data-logger/data"
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

type OrientationFeed struct {
	subscriptions data.Subscriptions
}

func NewOrientedAccelerationFeed() *OrientationFeed {
	return &OrientationFeed{
		subscriptions: make(data.Subscriptions),
	}
}

func (f *OrientationFeed) Subscribe(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	f.subscriptions[name] = sub
	return sub
}

var g = 0

var counter = 0
var lastOrientation = OrientationUnset

func (f *OrientationFeed) Start(subscription *data.Subscription) {
	go func() {

		// we assume a front orientation as a base orientation
		orientationCounter := make(OrientationCounter)
		//todo: stop lock for orientation when confident

		for {
			select {
			case event := <-subscription.IncomingEvents:
				g += 1
				if len(f.subscriptions) == 0 {
					continue
				}
				e := event.(*TiltCorrectedAccelerationEvent)
				newOrientation := computeOrientation(e.Acceleration.Acceleration)
				if newOrientation == OrientationUnset {
					lastOrientation = OrientationUnset
					counter = 0
					continue

				}

				if newOrientation != lastOrientation && lastOrientation != OrientationUnset {
					lastOrientation = newOrientation
					counter = 0
					continue
				}

				counter++
				//fmt.Println("Orientation:", newOrientation, lastOrientation, g)
				if counter > 20 {
					fmt.Println("Incrementing:", newOrientation)
					if newOrientation == OrientationFront {
						panic("should not happen")
					}
					orientationCounter.Increment(newOrientation)
				}

				if orientationCounter.Orientation() != OrientationUnset {
					a := NewAcceleration(e.Acceleration.Acceleration.X, e.Acceleration.Acceleration.Y, e.Acceleration.Acceleration.Z, e.Acceleration.Magnitude)
					a = FixAccelerationOrientation(a, orientationCounter.Orientation())
					t := FixTiltOrientation(e.Acceleration.TiltAngles, orientationCounter.Orientation())

					orientationEvent := NewOrientatedAccelerationEvent(NewOrientedAcceleration(a, t, orientationCounter.Orientation()))
					for _, sub := range f.subscriptions {
						sub.IncomingEvents <- orientationEvent
					}
				}

				lastOrientation = newOrientation
			}
		}
	}()
}

func computeOrientation(acceleration *Acceleration) Orientation {
	x := acceleration.X
	y := acceleration.Y

	fmt.Println("X:", x, "Y:", y)
	movementThreshold := 0.015

	if x > movementThreshold && y > -movementThreshold && y < movementThreshold {
		return OrientationFront
	} else if y < -movementThreshold && x > -movementThreshold && x < movementThreshold {
		return OrientationRight
	} else if y > movementThreshold && x > -movementThreshold && x < movementThreshold {
		return OrientationLeft
	} else if x < -movementThreshold && y > -movementThreshold && y < movementThreshold {
		return OrientationBack
	}

	return OrientationUnset
}

// OrientedAccelerationEvent X, Y and Z are the real world orientation values
// X forward and backwards, Y left and light and Z up and down
type OrientedAccelerationEvent struct {
	*data.BaseEvent
	Acceleration *OrientedAcceleration
}

func NewOrientatedAccelerationEvent(acceleration *OrientedAcceleration) *OrientedAccelerationEvent {
	orientationEvent := &OrientedAccelerationEvent{
		BaseEvent:    data.NewBaseEvent("OrientedAcceleration", "IMU"),
		Acceleration: acceleration,
	}
	return orientationEvent
}
