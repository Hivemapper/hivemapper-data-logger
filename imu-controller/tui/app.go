package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/streamingfast/hm-imu-logger/data"
	"github.com/streamingfast/hm-imu-logger/device/iim42652"
)

type App struct {
	sub *data.Subscription
	ui  *tea.Program
}

func NewApp(p *data.Pipeline) *App {
	sub := p.SubscribeAcceleration("tui")

	model := InitialModel(&iim42652.Acceleration{})
	ui := tea.NewProgram(model)

	return &App{
		sub: sub,
		ui:  ui,
	}
}

func (a *App) Run() (err error) {

	go func() {
		for {
			select {
			case acceleration := <-a.sub.IncomingAcceleration:
				a.ui.Send(acceleration)
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
