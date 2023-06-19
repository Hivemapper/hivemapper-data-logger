package imu

import (
	"fmt"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type TiltCorrectedAccelerationFeed struct {
	imu              *iim42652.IIM42652
	lastUpdate       interface{}
	xAngleCalibrated *data.AverageFloat64
	yAngleCalibrated *data.AverageFloat64
	zAngleCalibrated *data.AverageFloat64
	calibrated       bool
	handlers         []TiltCorrectedAccelerationHandler
}

type TiltCorrectedAccelerationHandler func(corrected *Acceleration, tiltAngles *TiltAngles, temperature iim42652.Temperature) error

func NewTiltCorrectedAccelerationFeed(handlers ...TiltCorrectedAccelerationHandler) *TiltCorrectedAccelerationFeed {
	f := &TiltCorrectedAccelerationFeed{
		xAngleCalibrated: data.NewAverageFloat64WithCount("angleX", 100),
		yAngleCalibrated: data.NewAverageFloat64WithCount("angleY", 100),
		zAngleCalibrated: data.NewAverageFloat64WithCount("angleZ", 100),
		handlers:         handlers,
	}

	return f
}

var continuousCount = 0
var xAvg = *data.NewAverageFloat64WithCount("", 30)
var yAvg = *data.NewAverageFloat64WithCount("", 30)
var zAvg = *data.NewAverageFloat64WithCount("", 30)

var afirst = true

func (f *TiltCorrectedAccelerationFeed) calibrate(acceleration *Acceleration) bool {
	magnitude := acceleration.Magnitude

	if afirst {
		afirst = false
		fmt.Println("first tilt handling", acceleration.Time)
	}

	if magnitude > 0.96 && magnitude < 1.04 {
		continuousCount++
		xAngle, yAngle, zAngle := computeTiltAngles(acceleration)
		xAvg.Add(xAngle)
		yAvg.Add(yAngle)
		zAvg.Add(zAngle)
		if continuousCount > 30 {
			f.xAngleCalibrated.Add(xAvg.Average)
			f.yAngleCalibrated.Add(yAvg.Average)
			f.zAngleCalibrated.Add(zAvg.Average)
			if !f.calibrated {
				fmt.Println("calibrated", f.xAngleCalibrated, f.yAngleCalibrated, f.zAngleCalibrated, acceleration.Time)
			}
			f.calibrated = true
		}
	} else {
		continuousCount = 0
		xAvg.Reset()
		yAvg.Reset()
		zAvg.Reset()
	}

	return f.calibrated
}

func (f *TiltCorrectedAccelerationFeed) HandleRawFeed(acceleration *Acceleration, _ *iim42652.AngularRate, temperature iim42652.Temperature) error {
	if !f.calibrate(acceleration) {
		return nil
	}

	correctedAcceleration := computeCorrectedGForce(acceleration, f.xAngleCalibrated.Average, f.yAngleCalibrated.Average, f.zAngleCalibrated.Average)
	angles := NewTiltAngles(f.xAngleCalibrated.Average, f.yAngleCalibrated.Average, f.zAngleCalibrated.Average)

	for _, handle := range f.handlers {
		err := handle(correctedAcceleration, angles, temperature)
		if err != nil {
			return fmt.Errorf("calling handler: %w", err)
		}
	}

	return nil
}
