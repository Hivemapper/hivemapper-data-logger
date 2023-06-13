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
	return orientation
}

type OrientationFeed struct {
	subscriptions data.Subscriptions
}

func NewOrientationFeed() *OrientationFeed {
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

var counter = 0
var lastOrientation = OrientationUnset

func (f *OrientationFeed) Start(subscription *data.Subscription) {
	go func() {

		// we assume a front orientation as a base orientation
		orientation := OrientationUnset
		orientationCounter := make(OrientationCounter)

		for {
			select {
			case event := <-subscription.IncomingEvents:
				if len(f.subscriptions) == 0 {
					continue
				}
				e := event.(*RawImuEvent)

				//if orientation == OrientationUnset {
				o := computeOrientation(e)
				if o == OrientationUnset || lastOrientation != o {
					lastOrientation = OrientationUnset
					counter = 0
					continue
				}

				counter++
				if counter > 20 {
					orientationCounter.Increment(o)
					fmt.Println("Mount Orientation:", orientation)
				}

				if orientationCounter.Orientation() != OrientationUnset {
					a := NewAcceleration(e.Acceleration.CamX(), e.Acceleration.CamY(), e.Acceleration.CamZ(), e.Acceleration.TotalMagnitude)
					a = FixAccelerationOrientation(a, orientationCounter.Orientation())

					orientationEvent := NewOrientatedAccelerationEvent(NewOrientedAcceleration(a, orientationCounter.Orientation()))
					for _, sub := range f.subscriptions {
						sub.IncomingEvents <- orientationEvent
					}
				}
			}
		}
	}()
}

func computeOrientation(event *RawImuEvent) Orientation {
	camX := event.Acceleration.CamX()
	camY := event.Acceleration.CamY()

	movementThreshold := 0.015

	if camX > movementThreshold && camY > -movementThreshold && camY < movementThreshold {
		return OrientationFront
	} else if camY < -movementThreshold && camX > -movementThreshold && camX < movementThreshold {
		return OrientationRight
	} else if camY > movementThreshold && camX > -movementThreshold && camX < movementThreshold {
		return OrientationLeft
	} else if camX < -movementThreshold && camY > -movementThreshold && camY < movementThreshold {
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
