package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/streamingfast/imu-controller/data"
)

type App struct {
	ui *tea.Program
}

func NewApp(p *data.EventEmitter) *App {
	model := InitialModel()
	ui := tea.NewProgram(model)
	app := &App{
		ui: ui,
	}

	p.RegisterEventHandler(app.HandleEvent)
	return app
}

func (a *App) HandleEvent(event data.Event) {
	msg := &MotionModelMsg{}
	switch event := event.(type) {
	case *data.ImuAccelerationEvent:
		msg.Acceleration = event.Acceleration
		msg.xAvg = event.AvgX
		msg.yAvg = event.AvgY
		msg.magnitudeAvg = event.AvgMagnitude
	case *data.StopEndEvent:
		msg.event = event.String()
	case *data.StopDetectEvent:
		msg.event = event.String()
	case *data.TurnEvent:
		msg.event = event.String()
	case *data.AccelerationEvent:
		msg.event = event.String()
	case *data.DecelerationEvent:
		msg.event = event.String()
	}
	a.ui.Send(msg)
}

func (a *App) Run() (err error) {
	if _, err = a.ui.Run(); err != nil {
		if err != tea.ErrProgramKilled {
			return err
		}
	}
	return nil
}
