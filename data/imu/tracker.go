package imu

import (
	"time"
)

type Tracker interface {
	trackAcceleration(lastUpdate time.Time, x float64, y float64, z float64)
}

type LeftTurnTracker struct {
	continuousCount int
	start           time.Time
	config          *Config
	emitFunc        emit
}

func (t *LeftTurnTracker) trackAcceleration(_ time.Time, x float64, y float64, _ float64) {
	magnitude := computeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y > t.config.LeftTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			t.emitFunc(NewLeftTurnEventDetected())
		}
	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.emitFunc(NewLeftTurnEvent(time.Since(t.start)))
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

func (t *RightTurnTracker) trackAcceleration(_ time.Time, x float64, y float64, _ float64) {
	magnitude := computeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y < t.config.RightTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			t.emitFunc(NewRightTurnEventDetected())
		}

	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.emitFunc(NewRightTurnEvent(time.Since(t.start)))
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

func (t *AccelerationTracker) trackAcceleration(lastUpdate time.Time, x float64, _ float64, _ float64) {
	if x > t.config.GForceAcceleratorThreshold {
		t.continuousCount++
		duration := time.Since(lastUpdate)
		t.speed += computeSpeedVariation(duration.Seconds(), x)
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.AccelerationContinuousCountWindow {
			t.emitFunc(NewAccelerationDetectedEvent())
		}

	} else {
		if t.continuousCount > t.config.AccelerationContinuousCountWindow {
			t.emitFunc(NewAccelerationEvent(t.speed, time.Since(t.start)))
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

func (t *DecelerationTracker) trackAcceleration(lastUpdate time.Time, x float64, _ float64, _ float64) {
	if x < t.config.GForceDeceleratorThreshold {
		t.continuousCount++
		duration := time.Since(lastUpdate)
		t.speed += computeSpeedVariation(duration.Seconds(), x)
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.DecelerationContinuousCountWindow {
			t.emitFunc(NewDecelerationDetectedEvent())
		}

	} else {
		if t.continuousCount > t.config.DecelerationContinuousCountWindow {
			t.emitFunc(NewDecelerationEvent(t.speed, time.Since(t.start)))
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

func (t *StopTracker) trackAcceleration(_ time.Time, x float64, y float64, z float64) {
	if x < 0.012 && x > -0.012 && y < 0.012 && y > -0.012 && z < 1.012 {
		t.continuousCount++

		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.StopEndContinuousCountWindow {
			t.emitFunc(NewStopDetectedEvent())
		}
	} else {
		if t.continuousCount > t.config.StopEndContinuousCountWindow {
			t.emitFunc(NewStopEndEvent(time.Since(t.start)))
		}

		t.continuousCount = 0
	}
}
