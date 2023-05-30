package data

import (
	"time"

	"github.com/streamingfast/hm-imu-logger/device/iim42652"
)

type Tracker interface {
	track(lastUpdate time.Time, acceleration *iim42652.Acceleration, xAvg *AverageFloat64, yAvg *AverageFloat64, totalMagnitudeAvg *AverageFloat64)
}

type LeftTurnTracker struct {
	continuousCount int
	start           time.Time
	emitFunc        emit
}

func (t *LeftTurnTracker) track(_ time.Time, _ *iim42652.Acceleration, _ *AverageFloat64, yAvg *AverageFloat64, totalMagnitudeAvg *AverageFloat64) {
	if totalMagnitudeAvg.Average > 0.2 && yAvg.Average > 0.15 {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
	} else {
		if t.continuousCount > 10 {
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
	emitFunc        emit
}

func (t *RightTurnTracker) track(_ time.Time, _ *iim42652.Acceleration, _ *AverageFloat64, yAvg *AverageFloat64, totalMagnitudeAvg *AverageFloat64) {
	if totalMagnitudeAvg.Average > 0.2 && yAvg.Average < -0.15 {
		t.continuousCount++
		if t.continuousCount == 1 {
			t.start = time.Now()
		}
	} else {
		if t.continuousCount > 10 {
			t.emitFunc(&TurnEvent{
				Direction: Right,
				Duration:  time.Since(t.start),
			})
		}
		t.continuousCount = 0
	}
}

type AccelerationTracker struct {
	emitFunc emit
}

func (t *AccelerationTracker) track(lastUpdate time.Time, acceleration *iim42652.Acceleration, xAvg *AverageFloat64, yAvg *AverageFloat64, totalMagnitudeAvg *AverageFloat64) {
	if xAvg.Average > 0 {
		duration := time.Since(lastUpdate)
		t.emitFunc(&AccelerationEvent{
			Speed:    computeSpeedVariation(duration.Seconds(), acceleration.CamX()),
			Duration: duration,
		})
	}
}

type DecelerationTracker struct {
	emitFunc emit
}

func (t *DecelerationTracker) track(lastUpdate time.Time, acceleration *iim42652.Acceleration, xAvg *AverageFloat64, yAvg *AverageFloat64, totalMagnitudeAvg *AverageFloat64) {
	if xAvg.Average < 0 {
		duration := time.Since(lastUpdate)
		t.emitFunc(&DecelerationEvent{
			Speed:    computeSpeedVariation(duration.Seconds(), acceleration.CamX()),
			Duration: duration,
		})
	}
}

type StopTracker struct {
	emitFunc emit
}

func (t *StopTracker) track(acceleration *iim42652.Acceleration, xAvg *AverageFloat64, yAvg *AverageFloat64, totalMagnitudeAvg *AverageFloat64) {

}
