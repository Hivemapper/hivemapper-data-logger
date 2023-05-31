package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
)

type App struct {
	ui                   *tea.Program
	imuEventSubscription *data.Subscription
}

func NewApp(imuEventSubscription *data.Subscription) *App {
	model := InitialModel()
	ui := tea.NewProgram(model)
	app := &App{
		ui:                   ui,
		imuEventSubscription: imuEventSubscription,
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
	case *imu.StopDetectEvent:
		msg.event = event.String()
	case *imu.TurnEvent:
		msg.event = event.String()
	case *imu.AccelerationEvent:
		msg.event = event.String()
	case *imu.DecelerationEvent:
		msg.event = event.String()
	}
	a.ui.Send(msg)
}

func (a *App) Run() (err error) {
	go func() {
		for {
			select {
			case event := <-a.imuEventSubscription.IncomingEvents:
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
