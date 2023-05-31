package imu

import (
	"time"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type Tracker interface {
	track(lastUpdate time.Time, acceleration *iim42652.Acceleration, xAvg *data.AverageFloat64, yAvg *data.AverageFloat64, totalMagnitudeAvg *data.AverageFloat64)
}

type LeftTurnTracker struct {
	continuousCount int
	start           time.Time
	config          *Config
	emitFunc        emit
}

func (t *LeftTurnTracker) track(_ time.Time, imuAccel *iim42652.Acceleration, _ *data.AverageFloat64, _ *data.AverageFloat64, _ *data.AverageFloat64) {
	magnitude := computeTotalMagnitude(imuAccel.CamX(), imuAccel.CamY())
	if magnitude > t.config.MinimumMagnitudeThreshold && imuAccel.CamY() > t.config.LeftTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
	} else {
		if t.continuousCount > t.config.ContinuousCountWindow {
			t.emitFunc(&TurnEvent{
				Direction: Left,
				Duration:  time.Since(t.start),
			})
		}
		t.continuousCount = 0
	}
}

type RightTurnTracker struct {
	continuousCount int
	start           time.Time
	config          *Config
	emitFunc        emit
}

func (t *RightTurnTracker) track(_ time.Time, imuAccel *iim42652.Acceleration, _ *data.AverageFloat64, _ *data.AverageFloat64, _ *data.AverageFloat64) {
	magnitude := computeTotalMagnitude(imuAccel.CamX(), imuAccel.CamY())
	if magnitude > t.config.MinimumMagnitudeThreshold && imuAccel.CamY() < t.config.RightTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
	} else {
		if t.continuousCount > t.config.ContinuousCountWindow {
			t.emitFunc(&TurnEvent{
				Direction: Right,
				Duration:  time.Since(t.start),
			})
		}
		t.continuousCount = 0
	}
}

type AccelerationTracker struct {
	continuousCount int
	speed           float64
	start           time.Time
	config          *Config
	emitFunc        emit
}

func (t *AccelerationTracker) track(lastUpdate time.Time, acceleration *iim42652.Acceleration, _ *data.AverageFloat64, _ *data.AverageFloat64, _ *data.AverageFloat64) {
	if acceleration.CamX() > t.config.GForceAcceleratorThreshold {
		t.continuousCount++
		duration := time.Since(lastUpdate)
		t.speed += computeSpeedVariation(duration.Seconds(), acceleration.CamX())
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
	} else {
		if t.continuousCount > t.config.ContinuousCountWindow {
			t.emitFunc(&AccelerationEvent{
				Speed:    t.speed,
				Duration: time.Since(t.start),
			})
		}
		t.speed = 0
		t.continuousCount = 0
	}
}

type DecelerationTracker struct {
	continuousCount int
	speed           float64
	start           time.Time
	config          *Config
	emitFunc        emit
}

func (t *DecelerationTracker) track(lastUpdate time.Time, acceleration *iim42652.Acceleration, _ *data.AverageFloat64, _ *data.AverageFloat64, _ *data.AverageFloat64) {
	if acceleration.CamX() < t.config.GForceDeceleratorThreshold {
		t.continuousCount++
		duration := time.Since(lastUpdate)
		t.speed += computeSpeedVariation(duration.Seconds(), acceleration.CamX())
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
	} else {
		if t.continuousCount > t.config.ContinuousCountWindow {
			t.emitFunc(&DecelerationEvent{
				Speed:    t.speed,
				Duration: time.Since(t.start),
			})
		}
		t.speed = 0
		t.continuousCount = 0
	}
}

type StopTracker struct {
	continuousCount int
	start           time.Time
	config          *Config
	emitFunc        emit
}

func (t *StopTracker) track(_ time.Time, acceleration *iim42652.Acceleration, _ *data.AverageFloat64, _ *data.AverageFloat64, _ *data.AverageFloat64) {
	if acceleration.CamX() == 0.0 && acceleration.CamY() == 0.0 && acceleration.CamZ() == 1.0 {
		t.continuousCount++

		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.StopEndContinuousCountWindow {
			t.emitFunc(&StopDetectEvent{})
		}
	} else {
		if t.continuousCount > t.config.StopEndContinuousCountWindow {
			t.emitFunc(&StopEndEvent{
				Duration: time.Since(t.start),
			})
		}

		t.continuousCount = 0
	}
}
