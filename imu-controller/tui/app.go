package tui

import (
	"math"
	"time"

	"github.com/streamingfast/hm-imu-logger/device/iim42652"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/streamingfast/hm-imu-logger/data"
)

type App struct {
	sub *data.Subscription
	ui  *tea.Program
}

func NewApp(p *data.Pipeline) *App {
	sub := p.SubscribeAcceleration("tui")

	model := InitialModel()
	ui := tea.NewProgram(model)

	return &App{
		sub: sub,
		ui:  ui,
	}
}

func (a *App) Run() (err error) {
	go func() {
		lastUpdate := time.Time{}
		var lastAcceleration iim42652.Acceleration
		xAvg := NewAverageFloat64("X average")
		yAvg := NewAverageFloat64("Y average")
		totalMagnitudeAvg := NewAverageFloat64("Total magnetude average")

		for {
			select {
			case acceleration := <-a.sub.IncomingAcceleration:
				if lastUpdate == (time.Time{}) {
					lastAcceleration = *acceleration
					lastUpdate = time.Now()
					continue
				}

				timeSinceLastUpdate := time.Since(lastUpdate)

				accelerationSpeed := computeAccelerationSpeed(timeSinceLastUpdate.Seconds(), lastAcceleration.CamX())
				xAvg.Add(lastAcceleration.CamX())
				yAvg.Add(lastAcceleration.CamY())
				totalMagnitudeAvg.Add(math.Sqrt(math.Pow(lastAcceleration.CamX(), 2) + math.Pow(lastAcceleration.CamY(), 2)))

				lastAcceleration = *acceleration

				motionModel := &MotionModelMsg{
					Acceleration:      acceleration,
					accelerationSpeed: &accelerationSpeed,
					xAvg:              xAvg,
					yAvg:              yAvg,
					totalMagnitudeAvg: totalMagnitudeAvg,
				}

				a.ui.Send(motionModel)
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
