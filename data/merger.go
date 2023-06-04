package data

type EventFeedMerger struct {
	subscriptions       Subscriptions
	mergedSubscriptions []*Subscription
	Events              chan Event
}

func NewEventFeedMerger(subscriptions ...*Subscription) *EventFeedMerger {
	return &EventFeedMerger{
		mergedSubscriptions: subscriptions,
		Events:              make(chan Event),
	}
}

func (m *EventFeedMerger) Subscribe(name string) *Subscription {
	sub := &Subscription{
		IncomingEvents: make(chan Event),
	}
	m.subscriptions[name] = sub
	return sub
}

func (m *EventFeedMerger) MergeEvents() {
	for _, sub := range m.mergedSubscriptions {
		go func(sub *Subscription) {
			for {
				select {
				case event := <-sub.IncomingEvents:
					m.Events <- event
				}
			}
		}(sub)
	}
}
