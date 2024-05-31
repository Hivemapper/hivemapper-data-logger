package gnss

import (
	"fmt"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/rosshemsley/kalman"
	"github.com/rosshemsley/kalman/models"
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
	g.lonModel = models.NewSimpleModel(d.Timestamp, d.Longitude, models.SimpleModelConfig{
		InitialVariance:     d.Longitude,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	g.lonFilter = kalman.NewKalmanFilter(g.lonModel)
	g.latModel = models.NewSimpleModel(d.Timestamp, d.Latitude, models.SimpleModelConfig{
		InitialVariance:     d.Latitude,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	g.latFilter = kalman.NewKalmanFilter(g.latModel)
	g.Data = d
}

type GnssFeed struct {
	dataHandler GnssDataHandler
}

func NewGnssFeed(dataHandler GnssDataHandler) *GnssFeed {
	g := &GnssFeed{
		dataHandler: dataHandler,
	}

	return g
}

func (f *GnssFeed) Run(gnssDevice *neom9n.Neom9n) error {
	//todo: datafeed is ugly
	dataFeed := neom9n.NewDataFeed(f.HandleData)
	err := gnssDevice.Run(dataFeed)
	if err != nil {
		return fmt.Errorf("running gnss device: %w", err)
	}
	return nil
}

func (f *GnssFeed) HandleData(d *neom9n.Data) {
	err := f.dataHandler(d)
	if err != nil {
		fmt.Printf("handling gnss data: %s\n", err)
	}

}
