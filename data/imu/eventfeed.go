package imu

import (
	"fmt"
	"os"
	"time"

	"github.com/rosshemsley/kalman"
	"github.com/rosshemsley/kalman/models"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type emit func(event data.Event)

type EventFeed struct {
	imu           *iim42652.IIM42652
	subscriptions data.Subscriptions
	config        *Config
}

func NewEventFeed(imu *iim42652.IIM42652, config *Config) *EventFeed {
	return &EventFeed{
		config:        config,
		imu:           imu,
		subscriptions: make(data.Subscriptions),
	}
}

func (p *EventFeed) Run() error {
	fmt.Println("Running pipeline")
	err := p.run()
	if err != nil {
		return fmt.Errorf("running pipeline: %w", err)
	}
	return nil
}

func (p *EventFeed) Subscribe(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	p.subscriptions[name] = sub
	return sub
}

func (p *EventFeed) emit(event data.Event) {
	event.SetTime(time.Now())
	for _, subscription := range p.subscriptions {
		subscription.IncomingEvents <- event
	}
}

func (p *EventFeed) run() error {
	now := time.Now()
	xModel := models.NewSimpleModel(now, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	xFilter := kalman.NewKalmanFilter(xModel)

	yModel := models.NewSimpleModel(now, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	yFilter := kalman.NewKalmanFilter(yModel)

	zModel := models.NewSimpleModel(now, 1.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	zFilter := kalman.NewKalmanFilter(zModel)

	leftTurnTracker := &LeftTurnTracker{
		config:   p.config,
		emitFunc: p.emit,
	}

	rightTurnTracker := &RightTurnTracker{
		config:   p.config,
		emitFunc: p.emit,
	}

	accelerationTracker := &AccelerationTracker{
		config:   p.config,
		emitFunc: p.emit,
	}

	decelerationTracker := &DecelerationTracker{
		config:   p.config,
		emitFunc: p.emit,
	}

	stopTracker := &StopTracker{
		config:   p.config,
		emitFunc: p.emit,
	}

	lastUpdate := time.Time{}

	for {
		acceleration, err := p.imu.GetAcceleration()
		if err != nil {
			panic(fmt.Errorf("getting acceleration: %w", err))
		}

		now := time.Now()
		err = xFilter.Update(now, xModel.NewMeasurement(acceleration.CamX()))
		if err != nil {
			return fmt.Errorf("updating kalman x filter: %w", err)
		}

		err = yFilter.Update(now, yModel.NewMeasurement(acceleration.CamY()))
		if err != nil {
			return fmt.Errorf("updating kalman y filter: %w", err)
		}

		err = zFilter.Update(now, zModel.NewMeasurement(acceleration.CamZ()))
		if err != nil {
			return fmt.Errorf("updating kalman z filter: %w", err)
		}

		x := xModel.Value(xFilter.State())
		y := yModel.Value(yFilter.State())
		z := zModel.Value(zFilter.State())

		magnitude := computeTotalMagnitude(acceleration.CamX(), acceleration.CamY())
		p.emit(NewImuAccelerationEvent(acceleration, x, y, z, magnitude))

		leftTurnTracker.trackAcceleration(lastUpdate, x, y, z)
		rightTurnTracker.trackAcceleration(lastUpdate, x, y, z)
		accelerationTracker.trackAcceleration(lastUpdate, x, y, z)
		decelerationTracker.trackAcceleration(lastUpdate, x, y, z)
		stopTracker.trackAcceleration(lastUpdate, x, y, z)

		//fmt.Println(acceleration.CamX(), ",", xAvg.Average, ",", xModel.Value(xFilter.State()))
		fmt.Fprintf(os.Stderr, "%f,%f,%f\n", x, y, z)

		lastUpdate = time.Now()

		time.Sleep(10 * time.Millisecond)
	}
}
