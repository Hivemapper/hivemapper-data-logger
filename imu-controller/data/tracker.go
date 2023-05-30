package data

import (
	"time"

	"github.com/streamingfast/hm-imu-logger/device/iim42652"
)

type tracker interface {
	track(acceleration *iim42652.Acceleration, xAvg *AverageFloat64, yAvg *AverageFloat64, totalMagnitudeAvg *AverageFloat64)
}

type LeftTurnTracker struct {
	continuousCount int
	start           time.Time
	emitFunc        emit
}

func (t *LeftTurnTracker) track(acceleration *iim42652.Acceleration, xAvg *AverageFloat64, yAvg *AverageFloat64, totalMagnitudeAvg *AverageFloat64) {
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
