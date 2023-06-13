package imu

import (
	"fmt"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type TiltCorrectedAccelerationEvent struct {
	*data.BaseEvent
	Acceleration *TiltCorrectedAcceleration
}

func NewTiltCorrectedAccelerationEvent(acceleration *OrientedAcceleration, xAngle, yAngle, zAngle float64) *TiltCorrectedAccelerationEvent {
	return &TiltCorrectedAccelerationEvent{
		BaseEvent:    data.NewBaseEvent("IMU_TILT_CORRECTED_ACCELERATION_EVENT", "IMU"),
		Acceleration: NewTiltCorrectedAcceleration(acceleration, xAngle, yAngle, zAngle),
	}
}

func (e *TiltCorrectedAccelerationEvent) String() string {
	return fmt.Sprintf("TiltCorrectedAccelerationEvent: %f, %f, Angles x %f, y %f, z %f", e.Acceleration.X, e.Acceleration.Y, e.Acceleration.XAngle, e.Acceleration.YAngle, e.Acceleration.ZAngle)
}

type TiltCorrectedAccelerationFeed struct {
	imu              *iim42652.IIM42652
	subscriptions    data.Subscriptions
	lastUpdate       interface{}
	xAngleCalibrated *data.AverageFloat64
	yAngleCalibrated *data.AverageFloat64
	zAngleCalibrated *data.AverageFloat64
	calibrated       bool
}

func NewTiltCorrectedAccelerationFeed() *TiltCorrectedAccelerationFeed {
	f := &TiltCorrectedAccelerationFeed{
		subscriptions:    make(data.Subscriptions),
		xAngleCalibrated: data.NewAverageFloat64WithCount("angleX", 100),
		yAngleCalibrated: data.NewAverageFloat64WithCount("angleY", 100),
		zAngleCalibrated: data.NewAverageFloat64WithCount("angleZ", 100),
	}

	return f
}

func (f *TiltCorrectedAccelerationFeed) Subscribe(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	f.subscriptions[name] = sub
	return sub
}

var continuousCount = 0
var xAvg = *data.NewAverageFloat64WithCount("", 30)
var yAvg = *data.NewAverageFloat64WithCount("", 30)
var zAvg = *data.NewAverageFloat64WithCount("", 30)

func (f *TiltCorrectedAccelerationFeed) calibrate(e *OrientedAccelerationEvent) bool {
	magnitude := e.Acceleration.Magnitude

	if magnitude > 0.96 && magnitude < 1.04 {
		continuousCount++
		xAngle, yAngle, zAngle := computeTiltAngles(e.Acceleration)
		xAvg.Add(xAngle)
		yAvg.Add(yAngle)
		zAvg.Add(zAngle)
	} else {
		if continuousCount > 30 {
			f.xAngleCalibrated.Add(xAvg.Average)
			f.yAngleCalibrated.Add(yAvg.Average)
			f.zAngleCalibrated.Add(zAvg.Average)
			fmt.Println("Calibrated:", f.xAngleCalibrated, f.yAngleCalibrated, f.zAngleCalibrated)
			f.calibrated = true
		}
		continuousCount = 0
		xAvg.Reset()
		yAvg.Reset()
		zAvg.Reset()
	}

	return f.calibrated
}

func (f *TiltCorrectedAccelerationFeed) Start(sub *data.Subscription) {
	fmt.Println("Running imu corrected feed")
	go func() {
		for {
			select {
			case event := <-sub.IncomingEvents:
				if len(f.subscriptions) == 0 {
					continue
				}

				e := event.(*OrientedAccelerationEvent)
				if !f.calibrate(e) {
					continue
				}

				correctedAcceleration := computeCorrectedGForce(e.Acceleration, f.xAngleCalibrated.Average, f.yAngleCalibrated.Average, f.zAngleCalibrated.Average)

				correctedEvent := NewTiltCorrectedAccelerationEvent(
					NewOrientedAcceleration(correctedAcceleration, e.Acceleration.Orientation),
					f.xAngleCalibrated.Average,
					f.yAngleCalibrated.Average,
					f.zAngleCalibrated.Average,
				)
				for _, subscription := range f.subscriptions {
					subscription.IncomingEvents <- correctedEvent
				}
			}
		}
	}()
}
