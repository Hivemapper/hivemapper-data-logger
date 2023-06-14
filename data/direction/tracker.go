package direction

import (
	"time"

	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
)

type Tracker interface {
	track(acceleration *imu.Acceleration, tiltAngles *imu.TiltAngles, orientation imu.Orientation, gnssData *neom9n.Data) data.Event
}

type LeftTurnTracker struct {
	continuousCount int
	start           time.Time
	config          *imu.Config
}

func (t *LeftTurnTracker) track(acceleration *imu.Acceleration, _ *imu.TiltAngles, _ imu.Orientation, _ *neom9n.Data) data.Event {
	x := acceleration.X
	y := acceleration.Y
	magnitude := imu.ComputeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y < t.config.LeftTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = acceleration.Time
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			return NewLeftTurnEventDetected(acceleration.Time)
		}
	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.continuousCount = 0
			return NewLeftTurnEvent(Since(t.start, acceleration.Time), acceleration.Time)
		}
	}
	return nil
}

type RightTurnTracker struct {
	continuousCount int
	start           time.Time
	config          *imu.Config
}

func (t *RightTurnTracker) track(acceleration *imu.Acceleration, _ *imu.TiltAngles, _ imu.Orientation, _ *neom9n.Data) data.Event {
	x := acceleration.X
	y := acceleration.Y
	magnitude := imu.ComputeTotalMagnitude(x, y)
	if magnitude > t.config.TurnMagnitudeThreshold && y > t.config.RightTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = acceleration.Time
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			return NewRightTurnEventDetected(acceleration.Time)
		}

	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.continuousCount = 0
			return NewRightTurnEvent(Since(t.start, acceleration.Time), acceleration.Time)
		}
	}
	return nil
}

type AccelerationTracker struct {
	continuousCount int
	speed           float64
	start           time.Time
	config          *imu.Config
}

func (t *AccelerationTracker) track(acceleration *imu.Acceleration, _ *imu.TiltAngles, _ imu.Orientation, _ *neom9n.Data) data.Event {
	x := acceleration.X
	if x > t.config.GForceAcceleratorThreshold {
		if t.continuousCount == 0 {
			t.start = acceleration.Time
			t.continuousCount++
			return nil
		}
		t.continuousCount++
		duration := Since(t.start, acceleration.Time)
		t.speed += imu.ComputeSpeedVariation(duration.Seconds(), x)

		if t.continuousCount == t.config.AccelerationContinuousCountWindow {
			return NewAccelerationDetectedEvent(acceleration.Time)
		}

	} else {
		if t.continuousCount > t.config.AccelerationContinuousCountWindow {
			t.continuousCount = 0
			t.speed = 0
			return NewAccelerationEvent(t.speed, time.Since(t.start), acceleration.Time)
		}
	}
	return nil
}

type DecelerationTracker struct {
	continuousCount int
	speed           float64
	start           time.Time
	config          *imu.Config
}

func (t *DecelerationTracker) track(acceleration *imu.Acceleration, _ *imu.TiltAngles, _ imu.Orientation, _ *neom9n.Data) data.Event {
	x := acceleration.X
	if x < t.config.GForceDeceleratorThreshold {
		if t.continuousCount == 0 {
			t.start = acceleration.Time
			t.continuousCount++
			return nil
		}
		t.continuousCount++
		duration := Since(t.start, acceleration.Time)
		t.speed += imu.ComputeSpeedVariation(duration.Seconds(), x)

		if t.continuousCount == t.config.DecelerationContinuousCountWindow {
			return NewDecelerationDetectedEvent(acceleration.Time)
		}

	} else {
		if t.continuousCount > t.config.DecelerationContinuousCountWindow {
			t.speed = 0
			t.continuousCount = 0
			return NewDecelerationEvent(t.speed, Since(t.start, acceleration.Time), acceleration.Time)
		}
	}
	return nil
}

type StopTracker struct {
	continuousCount int
	start           time.Time
	config          *imu.Config
}

func (t *StopTracker) track(acceleration *imu.Acceleration, _ *imu.TiltAngles, _ imu.Orientation, gnss *neom9n.Data) data.Event {
	if gnss == nil {
		return nil
	}
	if acceleration.Magnitude > 0.96 && acceleration.Magnitude < 1.04 && gnss.Speed > 1.5*3.6 {
		t.continuousCount++

		if t.continuousCount == 1 {
			t.start = acceleration.Time
		}
		if t.continuousCount == t.config.StopEndContinuousCountWindow {
			return NewStopDetectedEvent(acceleration.Time)
		}
	} else {
		if t.continuousCount > t.config.StopEndContinuousCountWindow {
			t.continuousCount = 0
			return NewStopEndEvent(Since(t.start, acceleration.Time), acceleration.Time)
		}

	}
	return nil
}

func Since(since time.Time, t time.Time) time.Duration {
	return t.Sub(since)
}
