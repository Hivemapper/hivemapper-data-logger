package direction

import (
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
)

type Tracker interface {
	track(acceleration *imu.Acceleration, tiltAngles *imu.TiltAngles, orientation imu.Orientation, gnssData *neom9n.Data) data.Event
}

type LeftTurnTracker struct {
	continuousCount int
	start           time.Time
	config          *imu.Config
	gnssData        *neom9n.Data
}

func (t *LeftTurnTracker) track(acceleration *imu.Acceleration, _ *imu.TiltAngles, _ imu.Orientation, gnssData *neom9n.Data) data.Event {
	y := acceleration.Y
	if y > t.config.LeftTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.gnssData = gnssData
			t.start = acceleration.Time
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			return NewLeftTurnEventDetected(acceleration.Time, t.gnssData)
		}
	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.continuousCount = 0
			//fmt.Println("Left turn ended:", Since(t.start, acceleration.Time))
			return NewLeftTurnEvent(Since(t.start, acceleration.Time), acceleration.Time, gnssData)
		}
	}
	return nil
}

type RightTurnTracker struct {
	continuousCount int
	start           time.Time
	gnssData        *neom9n.Data

	config *imu.Config
}

func (t *RightTurnTracker) track(acceleration *imu.Acceleration, _ *imu.TiltAngles, _ imu.Orientation, gnssData *neom9n.Data) data.Event {
	y := acceleration.Y
	if y < t.config.RightTurnThreshold {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.gnssData = gnssData
			t.start = acceleration.Time
		}
		if t.continuousCount == t.config.TurnContinuousCountWindow {
			return NewRightTurnEventDetected(t.start, t.gnssData)
		}

	} else {
		if t.continuousCount > t.config.TurnContinuousCountWindow {
			t.continuousCount = 0
			return NewRightTurnEvent(Since(t.start, acceleration.Time), acceleration.Time, gnssData)
		}
		t.continuousCount = 0
	}
	return nil
}

type AccelerationTracker struct {
	continuousCount int
	speed           float64
	start           time.Time
	gnssData        *neom9n.Data

	config *imu.Config
}

func (t *AccelerationTracker) track(acceleration *imu.Acceleration, _ *imu.TiltAngles, _ imu.Orientation, gnssData *neom9n.Data) data.Event {
	x := acceleration.X
	if x > t.config.GForceAcceleratorThreshold {
		if t.continuousCount == 0 {
			t.start = acceleration.Time
			t.gnssData = gnssData
			t.continuousCount++
			return nil
		}
		t.continuousCount++
		duration := Since(t.start, acceleration.Time)
		t.speed += imu.ComputeSpeedVariation(duration.Seconds(), x)

		if t.continuousCount == t.config.AccelerationContinuousCountWindow {
			return NewAccelerationDetectedEvent(t.start, t.gnssData)
		}

	} else {
		if t.continuousCount > t.config.AccelerationContinuousCountWindow {
			a := NewAccelerationEvent(t.speed, Since(t.start, acceleration.Time), acceleration.Time, gnssData)
			t.speed = 0
			t.continuousCount = 0
			return a
		}
		t.speed = 0
		t.continuousCount = 0
	}
	return nil
}

type DecelerationTracker struct {
	continuousCount int
	speed           float64
	start           time.Time
	gnssData        *neom9n.Data

	config *imu.Config
}

func (t *DecelerationTracker) track(acceleration *imu.Acceleration, _ *imu.TiltAngles, _ imu.Orientation, gnssData *neom9n.Data) data.Event {
	x := acceleration.X
	if x < t.config.GForceDeceleratorThreshold {
		if t.continuousCount == 0 {
			t.start = acceleration.Time
			t.gnssData = gnssData
			t.continuousCount++
			return nil
		}
		t.continuousCount++
		duration := Since(t.start, acceleration.Time)
		t.speed += imu.ComputeSpeedVariation(duration.Seconds(), x)

		if t.continuousCount == t.config.DecelerationContinuousCountWindow {
			return NewDecelerationDetectedEvent(t.start, t.gnssData)
		}

	} else {
		if t.continuousCount > t.config.DecelerationContinuousCountWindow {
			d := NewDecelerationEvent(t.speed, Since(t.start, acceleration.Time), acceleration.Time, gnssData)
			t.speed = 0
			t.continuousCount = 0
			return d
		}
		t.speed = 0
		t.continuousCount = 0
	}
	return nil
}

type StopTracker struct {
	continuousCount int
	start           time.Time
	gnssData        *neom9n.Data

	config *imu.Config
}

const km = 1

func (t *StopTracker) track(acceleration *imu.Acceleration, _ *imu.TiltAngles, _ imu.Orientation, gnssData *neom9n.Data) data.Event {
	if gnssData == nil {
		return nil
	}

	//fmt.Println("speed", gnssData.Speed*3.6, "mag", acceleration.Magnitude)

	if gnssData.Speed*3.6 < 4*km {
		t.continuousCount++

		if t.continuousCount == 1 {
			t.gnssData = gnssData
			t.start = acceleration.Time
		}
		if t.continuousCount == t.config.StopEndContinuousCountWindow {
			return NewStopDetectedEvent(t.start, t.gnssData)
		}
	} else {
		if t.continuousCount > t.config.StopEndContinuousCountWindow {
			t.continuousCount = 0
			return NewStopEndEvent(Since(t.start, acceleration.Time), acceleration.Time, gnssData)
		}

	}
	return nil
}

func Since(since time.Time, t time.Time) time.Duration {
	return t.Sub(since)
}
