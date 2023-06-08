package imu

import (
	"github.com/streamingfast/hivemapper-data-logger/data"
	"math"
)

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
		counter := 0
		lastOrientation := OrientationFront
		orientation := OrientationFront
		initialOrientationSet := false

		//frontOrientationCounter := 0
		//rightOrientationCounter := 0
		//leftOrientationCounter := 0
		//backOrientationCounter := 0

		for {
			select {
			case event := <-subscription.IncomingEvents:
				if len(f.subscriptions) == 0 {
					continue
				}
				e := event.(*RawImuEvent)
				ori := computeOrientation(e)

				if ori == lastOrientation {
					counter++
				} else {
					counter = 0
					lastOrientation = ori
				}

				if counter < 10 {
					// once we get 10 same orientations then we know that we have a correct orientation
					continue
				}

				if counter == 10 {
					orientation = lastOrientation
					initialOrientationSet = true
				}

				if counter == 1000 {
					orientation = lastOrientation
				}

				if initialOrientationSet {
					x, y, z := computeAxesOrientation(e)
					orientationEvent := NewOrientationEvent(x, y, z, e.Acceleration.TotalMagnitude, orientation)
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

	if camX > 0.05 && camY > -0.05 && camY < 0.05 {
		return OrientationFront
	} else if camY < 0.05 && camX > -0.05 && camX < 0.05 {
		return OrientationRight
	} else if camY > 0.05 && camX > -0.05 && camX < 0.05 {
		return OrientationLeft
	} else if camX < -0.05 && camY > -0.05 && camY < 0.05 {
		return OrientationBack
	}

	return OrientationUnset
}

func computeAxesOrientation(event *RawImuEvent) (float64, float64, float64) {
	return event.Acceleration.CamX(), event.Acceleration.CamY(), event.Acceleration.CamZ()
}

// OrientationEvent X, Y and Z are the real world orientation values
// X forward and backwards movement, Y left and light and Z up and down
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

func invert(val float64) float64 {
	return -val
}
