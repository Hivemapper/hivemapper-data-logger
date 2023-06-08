package imu

import (
	"fmt"
	"time"

	"github.com/rosshemsley/kalman"
	"github.com/rosshemsley/kalman/models"

	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type CorrectedAccelerationEvent struct {
	*data.BaseEvent
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	XAngle float64 `json:"x_angle"`
	YAngle float64 `json:"Y_angle"`
}

func (e *CorrectedAccelerationEvent) String() string {
	return fmt.Sprintf("CorrectedAccelerationEvent: %f, %f, Angles x %f, y %f", e.X, e.Y, e.XAngle, e.YAngle)
}

func NewCorrectedAccelerationEvent(x, y, xAngle, yAngle float64) *CorrectedAccelerationEvent {
	return &CorrectedAccelerationEvent{
		BaseEvent: data.NewBaseEvent("IMU_CORRECTED_ACCELERATION_EVENT", "IMU"),
		X:         x,
		Y:         y,
		XAngle:    xAngle,
		YAngle:    yAngle,
	}
}

type CorrectedAccelerationFeed struct {
	imu           *iim42652.IIM42652
	subscriptions data.Subscriptions
	lastUpdate    interface{}
	xModel        *models.SimpleModel
	xFilter       *kalman.KalmanFilter
	yModel        *models.SimpleModel
	yFilter       *kalman.KalmanFilter
}

func NewCorrectedAccelerationFeed() *CorrectedAccelerationFeed {
	f := &CorrectedAccelerationFeed{
		subscriptions: make(data.Subscriptions),
	}
	now := time.Now()
	f.lastUpdate = now
	f.xModel = models.NewSimpleModel(now, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	f.xFilter = kalman.NewKalmanFilter(f.xModel)

	f.yModel = models.NewSimpleModel(now, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	f.yFilter = kalman.NewKalmanFilter(f.yModel)

	return f
}

func (f *CorrectedAccelerationFeed) Subscribe(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	f.subscriptions[name] = sub
	return sub
}

func (f *CorrectedAccelerationFeed) Start(sub *data.Subscription) {
	fmt.Println("Running imu corrected feed")
	go func() {
		for {
			select {
			case event := <-sub.IncomingEvents:
				if len(f.subscriptions) == 0 {
					continue
				}
				e := event.(*OrientationEvent)
				x := e.GetX()
				y := e.GetY()
				z := e.GetZ()

				correctedX, correctedY := computeCorrectedGForce(x, y, z)
				xAngle, yAngle, _ := computeTiltAngles(correctedX, correctedY, 1)

				//now := time.Now()
				//err := f.xFilter.Update(now, f.xModel.NewMeasurement(correctedX))
				//if err != nil {
				//	panic(fmt.Errorf("updating x filter: %w", err))
				//}
				//x := f.xModel.Value(f.xFilter.State())
				//
				//err = f.yFilter.Update(now, f.yModel.NewMeasurement(correctedY))
				//if err != nil {
				//	panic(fmt.Errorf("updating y filter: %w", err))
				//}
				//y := f.yModel.Value(f.yFilter.State())

				correctedEvent := NewCorrectedAccelerationEvent(correctedX, correctedY, xAngle, yAngle)
				for _, subscription := range f.subscriptions {
					subscription.IncomingEvents <- correctedEvent
				}
			}
		}
	}()
}
