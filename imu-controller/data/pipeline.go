package data

import (
	"fmt"
	"time"

	"github.com/streamingfast/hm-imu-logger/device/iim42652"
)

type Subscription struct {
	IncomingAcceleration chan *iim42652.Acceleration
}

type subscriptions map[string]*Subscription

type Pipeline struct {
	imu           *iim42652.IIM42652
	subscriptions subscriptions
}

func NewPipeline(imu *iim42652.IIM42652) *Pipeline {
	return &Pipeline{
		imu:           imu,
		subscriptions: make(subscriptions),
	}
}

func (p *Pipeline) Run() error {
	err := p.run()
	if err != nil {
		return fmt.Errorf("running pipeline: %w", err)
	}
	return nil
}

func (p *Pipeline) SubscribeAcceleration(name string) *Subscription {
	sub := &Subscription{
		IncomingAcceleration: make(chan *iim42652.Acceleration),
	}
	p.subscriptions[name] = sub
	return sub
}

func (p *Pipeline) run() error {
	for {
		acceleration, err := p.imu.GetAcceleration()
		if err != nil {
			panic(fmt.Errorf("getting acceleration: %w", err))
		}
		for _, subscription := range p.subscriptions {
			subscription.IncomingAcceleration <- acceleration
		}
		time.Sleep(10 * time.Millisecond)
	}
}
