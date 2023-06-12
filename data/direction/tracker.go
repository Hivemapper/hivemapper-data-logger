package direction

import (
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"time"
)

type Tracker interface {
	track(e *imu.TiltCorrectedAccelerationEvent, g *gnss.GnssEvent)
}

type LeftTurnTracker struct {
	continuousCount int
	start           time.Time
	config          *imu.Config
	emitFunc        emit
}

func (t *LeftTurnTracker) track(e *imu.TiltCorrectedAccelerationEvent, g *gnss.GnssEvent) {
	x := e.X
	y := e.Y

	magnitude := imu.ComputeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y < t.config.LeftTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = e.GetTime()
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			t.emitFunc(NewLeftTurnEventDetected())
		}
	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.emitFunc(NewLeftTurnEvent(Since(t.start, e.GetTime())))
		}
		t.continuousCount = 0
	}
}

type RightTurnTracker struct {
	continuousCount int
	start           time.Time
	config          *imu.Config
	emitFunc        emit
}

func (t *RightTurnTracker) track(e *imu.TiltCorrectedAccelerationEvent, g *gnss.GnssEvent) {
	x := e.X
	y := e.Y
	magnitude := imu.ComputeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y > t.config.RightTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = e.GetTime()
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			t.emitFunc(NewRightTurnEventDetected())
		}

	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.emitFunc(NewRightTurnEvent(Since(t.start, e.GetTime())))
		}
		t.continuousCount = 0
	}
}

type AccelerationTracker struct {
	continuousCount int
	speed           float64
	start           time.Time
	config          *imu.Config
	emitFunc        emit
}

func (t *AccelerationTracker) track(e *imu.TiltCorrectedAccelerationEvent, g *gnss.GnssEvent) {
	x := e.X
	if x > t.config.GForceAcceleratorThreshold {
		if t.continuousCount == 0 {
			t.start = e.GetTime()
			t.continuousCount++
			return
		}
		t.continuousCount++
		duration := Since(t.start, e.GetTime())
		t.speed += imu.ComputeSpeedVariation(duration.Seconds(), x)

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
	config          *imu.Config
	emitFunc        emit
}

func (t *DecelerationTracker) track(e *imu.TiltCorrectedAccelerationEvent, g *gnss.GnssEvent) {
	x := e.X
	if x < t.config.GForceDeceleratorThreshold {
		if t.continuousCount == 0 {
			t.start = e.GetTime()
			t.continuousCount++
			return
		}
		t.continuousCount++
		duration := Since(t.start, e.GetTime())
		t.speed += imu.ComputeSpeedVariation(duration.Seconds(), x)

		if t.continuousCount == t.config.DecelerationContinuousCountWindow {
			t.emitFunc(NewDecelerationDetectedEvent())
		}

	} else {
		if t.continuousCount > t.config.DecelerationContinuousCountWindow {
			t.emitFunc(NewDecelerationEvent(t.speed, Since(t.start, e.GetTime())))
		}
		t.speed = 0
		t.continuousCount = 0
	}
}

type StopTracker struct {
	continuousCount int
	start           time.Time
	config          *imu.Config
	emitFunc        emit
}

func (t *StopTracker) track(e *imu.TiltCorrectedAccelerationEvent, g *gnss.GnssEvent) {
	if e.Magnitude > 0.96 && e.Magnitude < 1.04 && g.Data.Speed > 0.0 {
		t.continuousCount++

		if t.continuousCount == 1 {
			t.start = e.GetTime()
		}
		if t.continuousCount == t.config.StopEndContinuousCountWindow {
			t.emitFunc(NewStopDetectedEvent())
		}
	} else {
		if t.continuousCount > t.config.StopEndContinuousCountWindow {
			t.emitFunc(NewStopEndEvent(Since(t.start, e.GetTime())))
		}

		t.continuousCount = 0
	}
}

func Since(since time.Time, t time.Time) time.Duration {
	return t.Sub(since)
}
