package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/streamingfast/hm-imu-logger/data"
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
	case *data.TurnEvent:
	case *data.AccelerationEvent:
	case *data.DecelerationEvent:
	case *data.HeadingChangeEvent:
	case *data.StopEvent:
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
