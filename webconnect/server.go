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

type subscriptions map[string]*Subscription

type Subscription struct {
	IncomingEvents chan data.Event
	includes       []string
	excludes       []string
}

type GRPCEvent struct {
	*data.BaseEvent
	Response *eventsv1.EventsResponse
}

func NewGRPCEvent(resp *eventsv1.EventsResponse) *GRPCEvent {
	return &GRPCEvent{
		BaseEvent: data.NewBaseEvent("GRPC_EVENT"),
		Response:  resp,
	}
}

func (g *GRPCEvent) String() string {
	return "GRPCEvent"
}

type EventsServer struct {
	subscriptions subscriptions
	sync.Mutex
}

func NewEventServer(merger *data.EventFeedMerger) *EventsServer {
	es := &EventsServer{
		subscriptions: make(subscriptions),
	}

	fmt.Println("starting event server")
	merger.MergeEvents()

	go func() {
		for {
			select {
			case event := <-merger.Events:
				err := es.SendEvent(event)
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
	grpcEvent := NewGRPCEvent(&eventsv1.EventsResponse{
		Name:    event.GetName(),
		Payload: bytes,
	})

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

func (es *EventsServer) Subscribe(name string, includes []string, excludes []string) *Subscription {
	es.Lock()
	defer es.Unlock()

	sub := &Subscription{
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
