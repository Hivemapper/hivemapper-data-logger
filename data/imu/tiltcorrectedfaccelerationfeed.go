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

func NewTiltCorrectedAccelerationEvent(acceleration *Acceleration, tilt *TiltAngles) *TiltCorrectedAccelerationEvent {
	return &TiltCorrectedAccelerationEvent{
		BaseEvent:    data.NewBaseEvent("IMU_TILT_CORRECTED_ACCELERATION_EVENT", "IMU"),
		Acceleration: NewTiltCorrectedAcceleration(acceleration, tilt),
	}
}

func (e *TiltCorrectedAccelerationEvent) String() string {
	return fmt.Sprintf("TiltCorrectedAccelerationEvent")
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

func (f *TiltCorrectedAccelerationFeed) calibrate(acceleration *Acceleration) bool {
	magnitude := acceleration.Magnitude

	if magnitude > 0.96 && magnitude < 1.04 {
		continuousCount++
		xAngle, yAngle, zAngle := computeTiltAngles(acceleration)
		xAvg.Add(xAngle)
		yAvg.Add(yAngle)
		zAvg.Add(zAngle)
	} else {
		if continuousCount > 30 {
			f.xAngleCalibrated.Add(xAvg.Average)
			f.yAngleCalibrated.Add(yAvg.Average)
			f.zAngleCalibrated.Add(zAvg.Average)
			f.calibrated = true
		}
		continuousCount = 0
		xAvg.Reset()
		yAvg.Reset()
		zAvg.Reset()
	}

	//fmt.Println("Calibrated:", f.xAngleCalibrated, f.yAngleCalibrated, f.zAngleCalibrated, f.calibrated)
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

				e := event.(*RawImuEvent)
				if !f.calibrate(e.Acceleration) {
					continue
				}

				correctedAcceleration := computeCorrectedGForce(e.Acceleration, f.xAngleCalibrated.Average, f.yAngleCalibrated.Average, f.zAngleCalibrated.Average)

				correctedEvent := NewTiltCorrectedAccelerationEvent(
					correctedAcceleration,
					NewTiltAngles(f.xAngleCalibrated.Average, f.yAngleCalibrated.Average, f.zAngleCalibrated.Average),
				)
				for _, subscription := range f.subscriptions {
					subscription.IncomingEvents <- correctedEvent
				}
			}
		}
	}()
}
