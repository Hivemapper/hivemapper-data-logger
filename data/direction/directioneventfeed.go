package direction

import (
	"fmt"

	"github.com/streamingfast/imu-controller/device/iim42652"

	"github.com/rosshemsley/kalman"
	"github.com/rosshemsley/kalman/models"

	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
)

type DirectionEventHandler func(event data.Event) error

type DirectionEventFeed struct {
	config              *imu.Config
	leftTurnTracker     *LeftTurnTracker
	rightTurnTracker    *RightTurnTracker
	accelerationTracker *AccelerationTracker
	decelerationTracker *DecelerationTracker
	stopTracker         *StopTracker

	gnssData             *neom9n.Data
	handlers             []DirectionEventHandler
	filteredAcceleration *FilteredAcceleration
}

func NewDirectionEventFeed(config *imu.Config, handlers ...DirectionEventHandler) *DirectionEventFeed {
	feed := &DirectionEventFeed{
		config:               config,
		handlers:             handlers,
		filteredAcceleration: NewFilteredAcceleration(),
	}

	feed.leftTurnTracker = &LeftTurnTracker{
		config: config,
	}

	feed.rightTurnTracker = &RightTurnTracker{
		config: config,
	}

	feed.accelerationTracker = &AccelerationTracker{
		config: config,
	}

	feed.decelerationTracker = &DecelerationTracker{
		config: config,
	}

	feed.stopTracker = &StopTracker{
		config: config,
	}

	return feed
}

func (f *DirectionEventFeed) HandleGnssData(data *neom9n.Data) error {
	f.gnssData = data
	return nil
}

type FilteredAcceleration struct {
	*imu.Acceleration
	initialized bool
	xModel      *models.SimpleModel
	xFilter     *kalman.KalmanFilter
	yModel      *models.SimpleModel
	yFilter     *kalman.KalmanFilter
	zModel      *models.SimpleModel
	zFilter     *kalman.KalmanFilter
	mModel      *models.SimpleModel
	mFilter     *kalman.KalmanFilter
}

func NewFilteredAcceleration() *FilteredAcceleration {
	return &FilteredAcceleration{Acceleration: &imu.Acceleration{}}
}

func (f *FilteredAcceleration) init(acceleration *imu.Acceleration) {
	f.initialized = true

	f.xModel = models.NewSimpleModel(acceleration.Time, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	f.xFilter = kalman.NewKalmanFilter(f.xModel)

	f.yModel = models.NewSimpleModel(acceleration.Time, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	f.yFilter = kalman.NewKalmanFilter(f.yModel)

	f.zModel = models.NewSimpleModel(acceleration.Time, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	f.zFilter = kalman.NewKalmanFilter(f.zModel)

	f.mModel = models.NewSimpleModel(acceleration.Time, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	f.mFilter = kalman.NewKalmanFilter(f.mModel)

	f.Acceleration = acceleration
}

func (f *FilteredAcceleration) Update(acceleration *imu.Acceleration) (*imu.Acceleration, error) {
	err := f.xFilter.Update(acceleration.Time, f.xModel.NewMeasurement(acceleration.X))
	if err != nil {

		return nil, fmt.Errorf("updating x filter at %q: %w, ", acceleration.Time, err)
	}

	err = f.yFilter.Update(acceleration.Time, f.yModel.NewMeasurement(acceleration.Y))
	if err != nil {
		return nil, fmt.Errorf("updating y filter: %w", err)
	}

	err = f.zFilter.Update(acceleration.Time, f.zModel.NewMeasurement(acceleration.Z))
	if err != nil {
		return nil, fmt.Errorf("updating z filter: %w", err)
	}

	err = f.mFilter.Update(acceleration.Time, f.mModel.NewMeasurement(acceleration.Magnitude))
	if err != nil {
		return nil, fmt.Errorf("updating z filter: %w", err)
	}

	f.Acceleration.Time = acceleration.Time
	f.Acceleration.X = f.xModel.Value(f.xFilter.State())
	f.Acceleration.Y = f.yModel.Value(f.yFilter.State())
	f.Acceleration.Z = f.zModel.Value(f.zFilter.State())
	f.Acceleration.Magnitude = f.mModel.Value(f.mFilter.State())

	return f.Acceleration, nil
}

func (f *DirectionEventFeed) HandleOrientedAcceleration(acceleration *imu.Acceleration, tiltAngles *imu.TiltAngles, _ iim42652.Temperature, orientation imu.Orientation) error {
	if !f.filteredAcceleration.initialized {
		fmt.Println("initializing filtered acceleration", acceleration.Time)
		f.filteredAcceleration.init(acceleration)
	}

	updateAcceleration, err := f.filteredAcceleration.Update(acceleration)
	if err != nil {
		return fmt.Errorf("updating filtered acceleration: %w", err)
	}

	if e := f.leftTurnTracker.track(updateAcceleration, tiltAngles, orientation, f.gnssData); e != nil {
		if err := f.emit(e); err != nil {
			return fmt.Errorf("emitting left turn event: %w", err)
		}
	}

	if e := f.rightTurnTracker.track(updateAcceleration, tiltAngles, orientation, f.gnssData); e != nil {
		if err := f.emit(e); err != nil {
			return fmt.Errorf("emitting right turn event: %w", err)
		}
	}
	if e := f.accelerationTracker.track(updateAcceleration, tiltAngles, orientation, f.gnssData); e != nil {
		if err := f.emit(e); err != nil {
			return fmt.Errorf("emitting acceleration event: %w", err)
		}
	}
	if e := f.decelerationTracker.track(updateAcceleration, tiltAngles, orientation, f.gnssData); e != nil {
		if err := f.emit(e); err != nil {
			return fmt.Errorf("emitting deceleration event: %w", err)
		}
	}
	if e := f.stopTracker.track(updateAcceleration, tiltAngles, orientation, f.gnssData); e != nil {
		if err := f.emit(e); err != nil {
			return fmt.Errorf("emitting stop event: %w", err)
		}
	}

	return nil
}

func (f *DirectionEventFeed) emit(event data.Event) error {
	for _, handler := range f.handlers {
		err := handler(event)
		if err != nil {
			return fmt.Errorf("calling handler: %w", err)
		}
	}
	return nil
}
