package imu

import (
	"fmt"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"math"
)

type OrientationCounter struct {
	frontCounter int
	rightCounter int
	leftCounter  int
	backCounter  int
	unsetCounter int
}

func NewOrientationCounter() *OrientationCounter {
	return &OrientationCounter{
		frontCounter: 0,
		rightCounter: 0,
		leftCounter:  0,
		backCounter:  0,
		unsetCounter: 0,
	}
}

func (o *OrientationCounter) Reset() {
	o.frontCounter = 0
	o.rightCounter = 0
	o.leftCounter = 0
	o.backCounter = 0
	o.unsetCounter = 0
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

func (f *OrientationFeed) Start(subscription *data.Subscription) {
	//todo: as long as we don't know for sure what the direction of the mount is
	// then we won't be sending any events.
	// Once we know the direction and confidently know enough to estimate the direction:
	//  - read the last known direction in the sqlite and if we have something
	//    then
	// 	- return the event with directions AND then simply pass through the events
	//    no need to keep checking what is the direction that we have
	//  - write the direction in the sqlite

	go func() {
		initialOrientationSet := false

		// We assume a front orientation as a base orientation
		orientation := OrientationFront
		orientationCounter := NewOrientationCounter()

		for {
			select {
			case event := <-subscription.IncomingEvents:
				if len(f.subscriptions) == 0 {
					continue
				}
				e := event.(*RawImuEvent)

				if !initialOrientationSet {
					ori := computeOrientation(e)

					switch ori {
					case OrientationFront:
						orientationCounter.frontCounter++
						if orientationCounter.frontCounter > 50 {
							orientation = OrientationFront
							initialOrientationSet = true
							fmt.Println("Mount Orientation: Front ")
						}
					case OrientationRight:
						orientationCounter.rightCounter++
						if orientationCounter.rightCounter > 50 {
							orientation = OrientationRight
							initialOrientationSet = true
							fmt.Println("Mount Orientation: Right")
						}
					case OrientationLeft:
						orientationCounter.leftCounter++
						if orientationCounter.leftCounter > 50 {
							orientation = OrientationLeft
							initialOrientationSet = true
							fmt.Println("Mount Orientation: Left")
						}
					case OrientationBack:
						orientationCounter.backCounter++
						if orientationCounter.backCounter > 50 {
							orientation = OrientationBack
							initialOrientationSet = true
							fmt.Println("Mount Orientation: Back")
						}
					case OrientationUnset:
						orientationCounter.unsetCounter++
						if orientationCounter.unsetCounter > 50 {
							fmt.Println("Can't determine the mount direction, need to keep looping")
							orientationCounter.Reset()
						}
					}
				}

				if initialOrientationSet {
					orientationEvent := NewOrientationEvent(
						e.Acceleration.CamX(),
						e.Acceleration.CamY(),
						e.Acceleration.CamZ(),
						e.Acceleration.TotalMagnitude,
						orientation,
					)
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

// OrientationEvent X, Y and Z are the real world orientation values
// X forward and backwards, Y left and light and Z up and down
type OrientationEvent struct {
	*data.BaseEvent
	acceleration *OrientedAcceleration
}

func NewOrientationEvent(x, y, z, m float64, orientation Orientation) *OrientationEvent {
	return &OrientationEvent{
		BaseEvent:    data.NewBaseEvent("OrientedAcceleration", "IMU"),
		acceleration: NewOrientedAcceleration(NewBaseAcceleration(x, y, z, m), orientation),
	}
}

func (m *OrientationEvent) GetX() float64 {
	switch m.acceleration.GetOrientation() {
	case OrientationFront:
		return m.acceleration.GetX()
	case OrientationRight:
		return invert(m.acceleration.GetY())
	case OrientationLeft:
		return m.acceleration.GetY()
	case OrientationBack:
		return invert(m.acceleration.GetX())
	case OrientationUnset:
		return math.MaxFloat64
	default:
		panic("invalid orientation")
	}
}

func (m *OrientationEvent) GetY() float64 {
	switch m.acceleration.GetOrientation() {
	case OrientationFront:
		return m.acceleration.GetY()
	case OrientationRight:
		return m.acceleration.GetX()
	case OrientationLeft:
		return invert(m.acceleration.GetX())
	case OrientationBack:
		return invert(m.acceleration.GetY())
	case OrientationUnset:
		return math.MaxFloat64
	default:
		panic("invalid orientation")
	}
}

func (m *OrientationEvent) GetZ() float64 {
	return m.acceleration.GetZ()
}

func (m *OrientationEvent) GetOrientation() Orientation {
	return m.acceleration.GetOrientation()
}

func invert(val float64) float64 {
	return -val
}
