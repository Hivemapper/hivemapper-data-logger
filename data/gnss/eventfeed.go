package gnss

import (
	"fmt"
	"github.com/rosshemsley/kalman"
	"github.com/rosshemsley/kalman/models"
	"time"

	"github.com/streamingfast/gnss-controller/device/neom9n"
)

type GnssDataHandler func(data *neom9n.Data) error
type TimeHandler func(now time.Time) error

type GnssFilteredData struct {
	*neom9n.Data
	initialized bool
	lonModel    *models.SimpleModel
	lonFilter   *kalman.KalmanFilter
	latModel    *models.SimpleModel
	latFilter   *kalman.KalmanFilter
}

func NewGnssFilteredData() *GnssFilteredData {
	return &GnssFilteredData{Data: &neom9n.Data{}}
}

func (g *GnssFilteredData) init(d *neom9n.Data) {
	g.initialized = true
	g.lonModel = models.NewSimpleModel(d.Timestamp, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	g.lonFilter = kalman.NewKalmanFilter(g.lonModel)
	g.latModel = models.NewSimpleModel(d.Timestamp, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	g.latFilter = kalman.NewKalmanFilter(g.latModel)
	g.Data = d
}

type GnssFeed struct {
	dataHandlers     []GnssDataHandler
	timeHandlers     []TimeHandler
	gnssFilteredData *GnssFilteredData
}

func NewGnssFeed(dataHandlers []GnssDataHandler, timeHandlers []TimeHandler) *GnssFeed {
	return &GnssFeed{
		dataHandlers:     dataHandlers,
		timeHandlers:     timeHandlers,
		gnssFilteredData: NewGnssFilteredData(),
	}
}

func (f *GnssFeed) Run(gnssDevice *neom9n.Neom9n) error {
	//todo: datafeed is ugly
	dataFeed := neom9n.NewDataFeed(f.HandleData)
	err := gnssDevice.Run(dataFeed, func(now time.Time) {
		dataFeed.SetStartTime(now)
		for _, handler := range f.timeHandlers {
			err := handler(now)
			if err != nil {
				fmt.Printf("handling gnss time: %s\n", err)
			}
		}
	})
	if err != nil {
		return fmt.Errorf("running gnss device: %w", err)
	}

	return nil
}

func (f *GnssFeed) HandleData(d *neom9n.Data) {
	if !f.gnssFilteredData.initialized {
		f.gnssFilteredData.init(d)
	}

	filteredLon := d.Longitude
	filteredLat := d.Latitude

	//if d.Fix == "2D" || d.Fix == "3D" {
	err := f.gnssFilteredData.lonFilter.Update(d.Timestamp, f.gnssFilteredData.lonModel.NewMeasurement(d.Longitude))
	if err != nil {
		panic("updating lon filter: " + err.Error())
	}
	err = f.gnssFilteredData.latFilter.Update(d.Timestamp, f.gnssFilteredData.latModel.NewMeasurement(d.Latitude))
	if err != nil {
		panic("updating lat filter: " + err.Error())
	}

	filteredLon = f.gnssFilteredData.lonModel.Value(f.gnssFilteredData.lonFilter.State())
	filteredLat = f.gnssFilteredData.latModel.Value(f.gnssFilteredData.latFilter.State())
	//}

	d.Longitude = filteredLon
	d.Latitude = filteredLat

	for _, handler := range f.dataHandlers {
		err := handler(d)
		if err != nil {
			fmt.Printf("handling gnss data: %s\n", err)
		}
	}
}
