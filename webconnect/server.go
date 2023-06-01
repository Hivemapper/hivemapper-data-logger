package webconnect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bufbuild/connect-go"
	uuid2 "github.com/google/uuid"
	"github.com/streamingfast/hivemapper-data-logger/data"
	eventsv1 "github.com/streamingfast/hivemapper-data-logger/gen/proto/sf/events/v1"
	"sync"
)

type GRPCEvent struct {
	data.BaseEvent
	Response *eventsv1.EventsResponse
}

func (g *GRPCEvent) String() string {
	return "GRPCEvent"
}

type EventsServer struct {
	subscriptions data.Subscriptions
	sync.Mutex
}

func NewEventServer(imuEventSubscription *data.Subscription, gnssEventSubscription *data.Subscription) *EventsServer {
	es := &EventsServer{
		subscriptions: make(data.Subscriptions),
	}

	go func() {
		select {
		case imu := <-imuEventSubscription.IncomingEvents:
			err := es.SendEvent(imu)
			if err != nil {
				panic("failed to send event")
			}
		case gnss := <-gnssEventSubscription.IncomingEvents:
			err := es.SendEvent(gnss)
			if err != nil {
				panic("failed to send event")
			}
		}
	}()

	return es
}

func (es *EventsServer) SendEvent(event data.Event) error {
	es.Lock()
	defer es.Unlock()

	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling %s %w", event.GetName(), err)
	}

	grpcEvent := &GRPCEvent{
		Response: &eventsv1.EventsResponse{
			Name:    event.GetName(),
			Payload: bytes,
		},
	}

	for _, sub := range es.subscriptions {
		sub.IncomingEvents <- grpcEvent
	}

	return nil
}

func (es *EventsServer) Subscribe(name string) *data.Subscription {
	es.Lock()
	defer es.Unlock()

	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	es.subscriptions[name] = sub

	return sub
}

func (es *EventsServer) Unsubscribe(name string) {
	es.Lock()
	defer es.Unlock()

	delete(es.subscriptions, name)
}

func (es *EventsServer) Events(
	ctx context.Context,
	req *connect.Request[eventsv1.EventsRequest],
	stream *connect.ServerStream[eventsv1.EventsResponse],
) error {
	uuid := uuid2.NewString()
	sub := es.Subscribe(uuid)

	for {
		select {
		case <-ctx.Done():
			es.Unsubscribe(uuid)
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return nil
		case ev := <-sub.IncomingEvents:
			res := ev.(*GRPCEvent)

			err := stream.Send(res.Response)
			if err != nil {
				es.Unsubscribe(uuid)
				return fmt.Errorf("streaming: %w", err)
			}
		}
	}
}
