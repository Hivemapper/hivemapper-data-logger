package data

type EventFeedMerger struct {
	mergedSubscriptions []*Subscription
	Events              chan Event
}

func NewEventFeedMerger(subscriptions ...*Subscription) *EventFeedMerger {
	return &EventFeedMerger{
		mergedSubscriptions: subscriptions,
		Events:              make(chan Event),
	}
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
