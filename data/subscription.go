package data

type Subscriptions map[string]*Subscription

type Subscription struct {
	IncomingEvents chan Event
}
