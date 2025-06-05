package gnss

import (
	"fmt"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/gnss-controller/message"
)

type GnssDataHandler func(data *neom9n.Data) error
type TimeHandler func(now time.Time) error

type Option func(*GnssFeed)

type GnssFeed struct {
	dataHandlers []GnssDataHandler
	timeHandlers []TimeHandler

	skipFiltering bool
}

func NewGnssFeed(dataHandlers []GnssDataHandler, timeHandlers []TimeHandler, opts ...Option) *GnssFeed {
	g := &GnssFeed{
		dataHandlers: dataHandlers,
		timeHandlers: timeHandlers,
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

func WithSkipFiltering() func(*GnssFeed) {
	return func(f *GnssFeed) {
		f.skipFiltering = true
	}
}

func (f *GnssFeed) Run(gnssDevice *neom9n.Neom9n, redisFeed message.UbxMessageHandler, redisLogsEnabled bool) error {
	//todo: datafeed is ugly
	dataFeed := neom9n.NewDataFeed(f.HandleData)
	err := gnssDevice.Run(dataFeed, redisFeed, redisLogsEnabled)
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
