package imu

import (
	"github.com/streamingfast/hivemapper-data-logger/data"
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
	magnitude := computeTotalMagnitude(x, x)
	if magnitude > t.config.TurnMagnitudeThreshold && y > t.config.LeftTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.emitFunc(&TurnEvent{
				BaseEvent: &data.BaseEvent{
					Name: "LEFT_TURN_EVENT",
				},
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

func (t *RightTurnTracker) trackAcceleration(_ time.Time, x float64, y float64, _ float64) {
	magnitude := computeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y < t.config.RightTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.emitFunc(&TurnEvent{
				BaseEvent: &data.BaseEvent{
					Name: "RIGHT_TURN_EVENT",
				},
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

func (t *AccelerationTracker) trackAcceleration(lastUpdate time.Time, x float64, _ float64, _ float64) {
	if x > t.config.GForceAcceleratorThreshold {
		t.continuousCount++
		duration := time.Since(lastUpdate)
		t.speed += computeSpeedVariation(duration.Seconds(), x)
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.StopEndContinuousCountWindow {
			t.emitFunc(&AccelerationDetectedEvent{
				BaseEvent: &data.BaseEvent{
					Name: "ACCELERATION_DETECTED_EVENT",
				},
			})
		}

	} else {
		if t.continuousCount > t.config.AccelerationContinuousCountWindow {
			t.emitFunc(&AccelerationEvent{
				BaseEvent: &data.BaseEvent{
					Name: "ACCELERATION_EVENT",
				},
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

func (t *DecelerationTracker) trackAcceleration(lastUpdate time.Time, x float64, _ float64, _ float64) {
	if x < t.config.GForceDeceleratorThreshold {
		t.continuousCount++
		duration := time.Since(lastUpdate)
		t.speed += computeSpeedVariation(duration.Seconds(), x)
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
	} else {
		if t.continuousCount > t.config.DecelerationContinuousCountWindow {
			t.emitFunc(&DecelerationEvent{
				BaseEvent: &data.BaseEvent{
					Name: "DECELERATION_EVENT",
				},
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

func (t *StopTracker) trackAcceleration(_ time.Time, x float64, y float64, z float64) {
	if x < 0.012 && y < 0.012 && z < 1.012 {
		t.continuousCount++

		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.StopEndContinuousCountWindow {
			t.emitFunc(&StopDetectedEvent{
				BaseEvent: &data.BaseEvent{
					Name: "STOP_DETECTED_EVENT",
				},
			})
		}
	} else {
		if t.continuousCount > t.config.StopEndContinuousCountWindow {
			t.emitFunc(&StopEndEvent{
				BaseEvent: &data.BaseEvent{
					Name: "STOP_END_EVENT",
				},
				Duration: time.Since(t.start),
			})
		}

		t.continuousCount = 0
	}
}
