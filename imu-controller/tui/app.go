package tui

import (
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/streamingfast/hm-imu-logger/data"
)

type App struct {
	sub *data.Subscription
	ui  *tea.Program
}

func NewApp(p *data.AccelerationPipeline) *App {
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
		xAvg := NewAverageFloat64("X average")
		yAvg := NewAverageFloat64("Y average")
		totalMagnitudeAvg := NewAverageFloat64("Total magnetude average")

		for {
			select {
			case acceleration := <-a.sub.IncomingAcceleration:
				if lastUpdate == (time.Time{}) {
					lastUpdate = time.Now()
					continue
				}

				timeSinceLastUpdate := time.Since(lastUpdate)

				xAvg.Add(acceleration.CamX())
				yAvg.Add(acceleration.CamY())
				totalMagnitudeAvg.Add(math.Sqrt(math.Pow(acceleration.CamX(), 2) + math.Pow(acceleration.CamY(), 2)))

				speedVariation := computeSpeedVariation(timeSinceLastUpdate.Seconds(), acceleration.CamX())
				if math.Abs(speedVariation) < 0.02 {
					speedVariation = 0
				}
				speed := computeSpeed(speedVariation)

				motionModel := &MotionModelMsg{
					Acceleration:      acceleration,
					speed:             speed,
					speedVariation:    speedVariation,
					xAvg:              xAvg,
					yAvg:              yAvg,
					totalMagnitudeAvg: totalMagnitudeAvg,
				}

				a.ui.Send(motionModel)
				lastUpdate = time.Now()
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
