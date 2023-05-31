package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"os"
)

type App struct {
	ui                    *tea.Program
	imuEventSubscription  *data.Subscription
	gnssEventSubscription *data.Subscription
}

func NewApp(imuEventSubscription *data.Subscription, gnssEventSubscription *data.Subscription) *App {
	model := NewModel()
	ui := tea.NewProgram(model, tea.WithOutput(os.Stderr))
	app := &App{
		ui:                    ui,
		imuEventSubscription:  imuEventSubscription,
		gnssEventSubscription: gnssEventSubscription,
	}

	return app
}

func (a *App) HandleEvent(event data.Event) {
	msg := &MotionModelMsg{}
	switch event := event.(type) {
	case *imu.ImuAccelerationEvent:
		msg.Acceleration = event.Acceleration
		msg.xAvg = event.AvgX
		msg.yAvg = event.AvgY
		msg.magnitudeAvg = event.AvgMagnitude
	case *imu.StopEndEvent:
		msg.event = event.String()
	case *imu.StopDetectedEvent:
		msg.event = event.String()
	case *imu.TurnEvent:
		msg.event = event.String()
	case *imu.AccelerationEvent:
		msg.event = event.String()
	case *imu.DecelerationEvent:
		msg.event = event.String()
	case *gnss.GnssEvent:
		msg.gnssData = event.Data
	}
	a.ui.Send(msg)
}

func (a *App) Run() (err error) {
	go func() {
		for {
			select {
			case event := <-a.imuEventSubscription.IncomingEvents:
				a.HandleEvent(event)
			case event := <-a.gnssEventSubscription.IncomingEvents:
				a.HandleEvent(event)
			}
		}
	}()
	if _, err = a.ui.Run(); err != nil {
		if err != tea.ErrProgramKilled {
			return err
		}
	}
	return nil
}
