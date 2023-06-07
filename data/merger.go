package data

import (
	"fmt"
)

type EventFeedMerger struct {
	subscriptions       Subscriptions
	subscriptionToMerge []*Subscription
	Events              chan Event
}

func NewEventFeedMerger(subscriptions ...*Subscription) *EventFeedMerger {
	return &EventFeedMerger{
		subscriptions:       make(Subscriptions),
		subscriptionToMerge: subscriptions,
	}
}

func (m *EventFeedMerger) Subscribe(name string) *Subscription {
	sub := &Subscription{
		IncomingEvents: make(chan Event),
	}
	m.subscriptions[name] = sub
	return sub
}

func (m *EventFeedMerger) Start() {
	fmt.Println("Running event feed merger")
	fmt.Println("number of subscriptions to the event feed merger", len(m.subscriptions))
	fmt.Println("number of subscriptionsToMerge to the event feed merger", len(m.subscriptionToMerge))
	for _, sub := range m.subscriptionToMerge {
		go func(sub *Subscription) {
			for {
				select {
				case event := <-sub.IncomingEvents:
					if len(m.subscriptions) == 0 {
						continue
					}
					for _, sub := range m.subscriptions {
						sub.IncomingEvents <- event
					}
				}
			}
		}(sub)
	}
}
