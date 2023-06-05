package imu

import (
	"fmt"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"time"
)

type Tracker interface {
	trackAcceleration(lastUpdate time.Time, xAvg float64, yAvg float64, zAvg float64)
}

type LeftTurnTracker struct {
	continuousCount int
	start           time.Time
	config          *Config
	emitFunc        emit
}

func (t *LeftTurnTracker) trackAcceleration(_ time.Time, x float64, y float64, z float64) {
	magnitude := computeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y > t.config.LeftTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			t.emitFunc(NewLeftTurnEventDetected(data.NewGForcePosition(x, y, z)))
		}
	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.emitFunc(NewLeftTurnEvent(time.Since(t.start), data.NewGForcePosition(x, y, z)))
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

func (t *RightTurnTracker) trackAcceleration(_ time.Time, x float64, y float64, z float64) {
	magnitude := computeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y < t.config.RightTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			t.emitFunc(NewRightTurnEventDetected(data.NewGForcePosition(x, y, z)))
		}

	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.emitFunc(NewRightTurnEvent(time.Since(t.start), data.NewGForcePosition(x, y, z)))
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

func (t *AccelerationTracker) trackAcceleration(lastUpdate time.Time, x float64, y float64, z float64) {
	if x > t.config.GForceAcceleratorThreshold {
		t.continuousCount++
		duration := time.Since(lastUpdate)
		t.speed += computeSpeedVariation(duration.Seconds(), x)
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.AccelerationContinuousCountWindow {
			t.emitFunc(NewAccelerationDetectedEvent(data.NewGForcePosition(x, y, z)))
		}

	} else {
		if t.continuousCount > t.config.AccelerationContinuousCountWindow {
			t.emitFunc(NewAccelerationEvent(t.speed, time.Since(t.start), data.NewGForcePosition(x, y, z)))
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

func (t *DecelerationTracker) trackAcceleration(lastUpdate time.Time, x float64, y float64, z float64) {
	if x < t.config.GForceDeceleratorThreshold {
		t.continuousCount++
		duration := time.Since(lastUpdate)
		t.speed += computeSpeedVariation(duration.Seconds(), x)
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
		if t.continuousCount == t.config.DecelerationContinuousCountWindow {
			t.emitFunc(NewDecelerationDetectedEvent(data.NewGForcePosition(x, y, z)))
		}

	} else {
		if t.continuousCount > t.config.DecelerationContinuousCountWindow {
			t.emitFunc(NewDecelerationEvent(t.speed, time.Since(t.start), data.NewGForcePosition(x, y, z)))
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
			t.emitFunc(NewStopDetectedEvent(data.NewGForcePosition(x, y, z)))
		}
	} else {
		if t.continuousCount > t.config.StopEndContinuousCountWindow {
			t.emitFunc(NewStopEndEvent(time.Since(t.start), data.NewGForcePosition(x, y, z)))
		}

		t.continuousCount = 0
	}
}

type CorrectDataTracker struct {
	continuousCount int
	start           time.Time
	config          *Config
	emitFunc        emit
}

func (c *CorrectDataTracker) trackAcceleration(_ time.Time, x, y, z float64) {
	//c.continuousCount++
	//
	//if c.continuousCount == 1 {
	//	c.start = time.Now()
	//}

	// every 10 counts, we send a CorrectedEvent
	//if c.continuousCount > 10 && z != 1.00 {
	correctedX := computeCorrectedGForce(z, x)
	correctedY := computeCorrectedGForce(z, y)
	tiltAngleXDegrees := computeTiltAngle(z, x)
	tiltAngleYDegrees := computeTiltAngle(z, y)

	correctedZ := z // todo: not sure about this value
	tiltAngleZDegrees := 90 - tiltAngleXDegrees

	//fmt.Println("correctedX", correctedX)
	//fmt.Println("tilt angle X", tiltAngleXDegrees)
	//
	//fmt.Println("correctedY", correctedY)
	//fmt.Println("tilt angle Y", tiltAngleYDegrees)
	//
	//fmt.Println("correctedZ", correctedZ)
	//fmt.Println("tilt angle Z", tiltAngleZDegrees)

	angles := data.NewAngles(tiltAngleXDegrees, tiltAngleYDegrees, tiltAngleZDegrees)
	correctedGForceValues := data.NewGForcePosition(correctedX, correctedY, correctedZ)
	originalGForceValues := data.NewGForcePosition(x, y, z)

	correctedEvent := NewCorrectedEvent(time.Since(c.start), originalGForceValues, correctedGForceValues, angles)
	fmt.Println(correctedEvent.String())

	c.emitFunc(correctedEvent)
	c.continuousCount = 0
	//}
}
