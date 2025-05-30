package webconnect

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/imu-controller/device/iim42652"

	"github.com/Hivemapper/hivemapper-data-logger/data"
	eventsv1 "github.com/Hivemapper/hivemapper-data-logger/gen/proto/sf/events/v1"
	"github.com/bufbuild/connect-go"
	uuid2 "github.com/google/uuid"
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
		BaseEvent: data.NewBaseEvent("GRPC_EVENT", "GRPC", time.Now().UTC(), nil),
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

func NewEventServer() *EventsServer {
	es := &EventsServer{
		subscriptions: make(subscriptions),
	}

	return es
}

func (s *EventsServer) HandleOrientedAcceleration(corrected *imu.Acceleration, tiltAngles *imu.TiltAngles, orientation imu.Orientation) error {
	panic("implement me")
}

func (s *EventsServer) HandleGnssData(gnssData *neom9n.Data) error {
	event := data.NewBaseEvent("GNSS_EVENT", "GNSS", time.Now().UTC(), gnssData)
	err := s.SendEvent(event)
	if err != nil {
		return fmt.Errorf("sending event %w", err)
	}
	return nil
}

func (s *EventsServer) HandleRawImuFeed(acceleration *imu.Acceleration, angularRate *iim42652.AngularRate) error {
	panic("implement me")
}

func (s *EventsServer) HandleDirectionEvent(event data.Event) error {
	err := s.SendEvent(event)
	if err != nil {
		return fmt.Errorf("sending event %w", err)
	}
	return nil
}

func (s *EventsServer) SendEvent(event data.Event) error {
	s.Lock()
	defer s.Unlock()
	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling %s %w", event.GetName(), err)
	}
	grpcEvent := NewGRPCEvent(&eventsv1.EventsResponse{
		Name:    event.GetName(),
		Payload: bytes,
	})

	for _, sub := range s.subscriptions {
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

func (s *EventsServer) Subscribe(name string, includes []string, excludes []string) *Subscription {
	s.Lock()
	defer s.Unlock()

	sub := &Subscription{
		IncomingEvents: make(chan data.Event),
		includes:       includes,
		excludes:       excludes,
	}
	s.subscriptions[name] = sub
	fmt.Println("subscribed", name)
	return sub
}

func (s *EventsServer) Unsubscribe(name string) {
	s.Lock()
	defer s.Unlock()

	delete(s.subscriptions, name)
}

func (s *EventsServer) Events(
	ctx context.Context,
	req *connect.Request[eventsv1.EventsRequest],
	stream *connect.ServerStream[eventsv1.EventsResponse],
) error {
	uuid := uuid2.NewString()
	fmt.Println("request received", uuid, req.Msg.Includes, req.Msg.Excludes)
	sub := s.Subscribe(uuid, req.Msg.Includes, req.Msg.Excludes)

	for {
		select {
		case <-ctx.Done():
			s.Unsubscribe(uuid)
			if ctx.Err() != nil {
				fmt.Println("context err:", ctx.Err())
				return ctx.Err()
			}
			return nil
		case ev := <-sub.IncomingEvents:
			res := ev.(*GRPCEvent)
			err := stream.Send(res.Response)
			if err != nil {
				s.Unsubscribe(uuid)
				return fmt.Errorf("streaming: %w", err)
			}
		}
	}
}
