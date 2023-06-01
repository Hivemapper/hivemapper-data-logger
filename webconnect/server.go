package webconnect

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/bufbuild/connect-go"
	uuid2 "github.com/google/uuid"
	"github.com/streamingfast/hivemapper-data-logger/data"
	eventsv1 "github.com/streamingfast/hivemapper-data-logger/gen/proto/sf/events/v1"
)

type subscriptions map[string]*subscription

type subscription struct {
	IncomingEvents chan data.Event
	includes       []string
	excludes       []string
}

type GRPCEvent struct {
	data.BaseEvent
	Response *eventsv1.EventsResponse
}

func (g *GRPCEvent) String() string {
	return "GRPCEvent"
}

type EventsServer struct {
	subscriptions subscriptions
	sync.Mutex
}

func NewEventServer(imuEventSubscription *data.Subscription, gnssEventSubscription *data.Subscription) *EventsServer {
	es := &EventsServer{
		subscriptions: make(subscriptions),
	}

	fmt.Println("starting event server")
	go func() {
		for {
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
		send := true
		if len(sub.includes) > 0 {
			for _, include := range sub.includes {
				if include == event.GetName() {
					send = true
					break
				}
			}
		}
		if len(sub.excludes) > 0 {
			for _, exclude := range sub.excludes {
				if exclude == event.GetName() {
					send = false
					break
				}
			}
		}
		if send {
			sub.IncomingEvents <- grpcEvent
		}
	}

	return nil
}

func (es *EventsServer) Subscribe(name string, includes []string, excludes []string) *subscription {
	es.Lock()
	defer es.Unlock()

	sub := &subscription{
		IncomingEvents: make(chan data.Event),
		includes:       includes,
		excludes:       excludes,
	}
	es.subscriptions[name] = sub
	fmt.Println("subscribed", name)
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
	fmt.Println("request received", uuid, req.Msg.Includes, req.Msg.Excludes)
	sub := es.Subscribe(uuid, req.Msg.Includes, req.Msg.Excludes)

	for {
		select {
		case <-ctx.Done():
			es.Unsubscribe(uuid)
			if ctx.Err() != nil {
				fmt.Println("context err:", ctx.Err())
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
