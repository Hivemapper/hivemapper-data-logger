package gnss

import (
	"fmt"
	"time"

	"github.com/streamingfast/gnss-controller/device/neom9n"
)

type GnssDataHandler func(data *neom9n.Data) error
type TimeHandler func(now time.Time) error

type GnssFeed struct {
	dataHandlers []GnssDataHandler
	timeHandlers []TimeHandler
}

func NewGnssFeed(dataHandlers []GnssDataHandler, timeHandlers []TimeHandler) *GnssFeed {
	return &GnssFeed{
		dataHandlers: dataHandlers,
		timeHandlers: timeHandlers,
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
	for _, handler := range f.dataHandlers {
		err := handler(d)
		if err != nil {
			fmt.Printf("handling gnss data: %s\n", err)
		}
	}
}
