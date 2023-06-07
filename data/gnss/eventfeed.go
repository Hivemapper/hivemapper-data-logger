package gnss

import (
	"fmt"
	"strings"
	"time"

	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
)

type GnssEvent struct {
	*data.BaseEvent
	Data *neom9n.Data `json:"data"`
}

func NewGnssEvent(d *neom9n.Data) *GnssEvent {
	return &GnssEvent{
		BaseEvent: data.NewBaseEvent("GNSS_EVENT", "GNSS"), // no x, y, z GForces for gnss data
		Data:      d,
	}
}

func (g *GnssEvent) String() string {
	var sb strings.Builder
	sb.WriteString("GNSS:")
	sb.WriteString(fmt.Sprintf(" Latitude: %.5f", g.Data.Latitude))
	sb.WriteString(fmt.Sprintf(" Longitude: %.5f", g.Data.Longitude))
	sb.WriteString(fmt.Sprintf(" Altitude: %.2f ", g.Data.Altitude))
	sb.WriteString(fmt.Sprintf(" Heading: %.2f", g.Data.Heading))
	sb.WriteString(fmt.Sprintf(" Speed: %.2f", g.Data.Speed))
	return sb.String()
}

type GnssTimeSetEvent struct {
	*data.BaseEvent
	Time time.Time `json:"time"`
}

func NewGnssTimeSetEvent(t time.Time) *GnssTimeSetEvent {
	return &GnssTimeSetEvent{
		BaseEvent: data.NewBaseEvent("GNSS_TIME_SET_EVENT", "GNSS"),
		Time:      t,
	}
}

func (g *GnssTimeSetEvent) String() string {
	return fmt.Sprintf("GNSS_TIME_SET_EVENT: %s", g.Time)
}

type EventFeed struct {
	subscriptions data.Subscriptions
}

func NewEventFeed() *EventFeed {
	return &EventFeed{
		subscriptions: make(data.Subscriptions),
	}
}

func (f *EventFeed) Subscribe(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	f.subscriptions[name] = sub
	return sub
}

func (f *EventFeed) Start(gnssDevice *neom9n.Neom9n) {
	fmt.Println("Running gnss feed")
	go func() {
		//todo: datafeed is ugly
		dataFeed := neom9n.NewDataFeed(f.HandleData)
		err := gnssDevice.Run(dataFeed, func(now time.Time) {
			dataFeed.SetStartTime(now)
			f.emit(NewGnssTimeSetEvent(now))
		})
		if err != nil {
			panic(fmt.Errorf("running gnss device: %w", err))
		}
	}()
}

func (f *EventFeed) HandleData(d *neom9n.Data) {
	f.emit(NewGnssEvent(d))
}

func (f *EventFeed) emit(event data.Event) {
	for _, subscription := range f.subscriptions {
		subscription.IncomingEvents <- event
	}
}
