package data

import (
	"fmt"
	"sync"
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

func (m *EventFeedMerger) Run() {
	fmt.Println("Running event feed merger")
	wg := sync.WaitGroup{}
	for _, sub := range m.subscriptionToMerge {
		wg.Add(1)
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
			wg.Done()
		}(sub)
	}
	wg.Wait()
}
