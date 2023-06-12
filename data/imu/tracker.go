package imu

import (
	"time"
)

type Tracker interface {
	track(lastUpdate time.Time, e *TiltCorrectedAccelerationEvent)
}

type LeftTurnTracker struct {
	continuousCount int
	start           time.Time
	config          *Config
	emitFunc        emit
}

func (t *LeftTurnTracker) track(now time.Time, e *TiltCorrectedAccelerationEvent) {
	x := e.X
	y := e.Y

	magnitude := computeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y < t.config.LeftTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = now
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			t.emitFunc(NewLeftTurnEventDetected())
		}
	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.emitFunc(NewLeftTurnEvent(Since(t.start, now)))
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

func (t *RightTurnTracker) track(now time.Time, e *TiltCorrectedAccelerationEvent) {
	x := e.X
	y := e.Y
	magnitude := computeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y > t.config.RightTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = now
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			t.emitFunc(NewRightTurnEventDetected())
		}

	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.emitFunc(NewRightTurnEvent(Since(t.start, now)))
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

func (t *AccelerationTracker) track(now time.Time, e *TiltCorrectedAccelerationEvent) {
	x := e.X
	if x > t.config.GForceAcceleratorThreshold {
		if t.continuousCount == 0 {
			t.start = now
			t.continuousCount++
			return
		}
		t.continuousCount++
		duration := Since(t.start, now)
		t.speed += computeSpeedVariation(duration.Seconds(), x)

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

func (t *DecelerationTracker) track(now time.Time, e *TiltCorrectedAccelerationEvent) {
	x := e.X
	if x < t.config.GForceDeceleratorThreshold {
		if t.continuousCount == 0 {
			t.start = now
			t.continuousCount++
			return
		}
		t.continuousCount++
		duration := Since(t.start, now)
		t.speed += computeSpeedVariation(duration.Seconds(), x)

		if t.continuousCount == t.config.DecelerationContinuousCountWindow {
			t.emitFunc(NewDecelerationDetectedEvent())
		}

	} else {
		if t.continuousCount > t.config.DecelerationContinuousCountWindow {
			t.emitFunc(NewDecelerationEvent(t.speed, Since(t.start, now)))
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

func (t *StopTracker) track(now time.Time, e *TiltCorrectedAccelerationEvent) {
	//todo: we need gps speed
	if e.Magnitude > 0.96 && e.Magnitude < 1.04 {
		t.continuousCount++

		if t.continuousCount == 1 {
			t.start = now
		}
		if t.continuousCount == t.config.StopEndContinuousCountWindow {
			t.emitFunc(NewStopDetectedEvent())
		}
	} else {
		if t.continuousCount > t.config.StopEndContinuousCountWindow {
			t.emitFunc(NewStopEndEvent(Since(t.start, now)))
		}

		t.continuousCount = 0
	}
}

func Since(since time.Time, t time.Time) time.Duration {

	return t.Sub(since)
}
