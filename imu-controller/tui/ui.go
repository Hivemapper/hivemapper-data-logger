package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/streamingfast/hm-imu-logger/device/iim42652"
)

const GraphHeader = " |                                                                                                       | \n"
const GraphFooter = "-1                                                   0                                                   1 \n"

type MotionModel struct {
	Acceleration     *iim42652.Acceleration
	speed            float64
	averageMagnitude float64
	xAvg             float64
	yAvg             float64
}

type Model struct {
	MotionModel *MotionModel
}

type MotionModelMsg struct {
	Acceleration      *iim42652.Acceleration
	speed             *float64
	xAvg              *AverageFloat64
	yAvg              *AverageFloat64
	totalMagnitudeAvg *AverageFloat64
}

func InitialModel() Model {
	return Model{MotionModel: &MotionModel{Acceleration: &iim42652.Acceleration{}, speed: 0.0, averageMagnitude: 0.0}}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case *MotionModelMsg:
		if msg.Acceleration != nil {
			m.MotionModel.Acceleration = msg.Acceleration
		}
		if msg.speed != nil {
			m.MotionModel.speed = *msg.speed
		}
		if msg.xAvg != nil {
			m.MotionModel.xAvg = msg.xAvg.Average
		}
		if msg.yAvg != nil {
			m.MotionModel.yAvg = msg.yAvg.Average
		}
		if msg.totalMagnitudeAvg != nil {
			m.MotionModel.averageMagnitude = msg.totalMagnitudeAvg.Average
		}

	}

	return m, nil
}

func (m Model) View() string {
	var graph strings.Builder
	graph.WriteString(GraphHeader)

	var graphBody strings.Builder

	graphBody.WriteString(createAxisGString(m.MotionModel.Acceleration.CamX(), "X"))
	graphBody.WriteString(createAxisGString(m.MotionModel.Acceleration.CamY(), "Y"))
	graphBody.WriteString(createAxisGString(m.MotionModel.Acceleration.CamZ()-1, "Z"))

	graph.WriteString(graphBody.String())
	graph.WriteString(GraphFooter)

	graph.WriteString("\n")
	if m.MotionModel.speed > 0.0 {
		graph.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Render(fmt.Sprintf("SPEED: %.2f \n", m.MotionModel.speed)))
	} else if m.MotionModel.speed < 0.0 {
		graph.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(fmt.Sprintf("SPEED: %.2f \n", m.MotionModel.speed)))
	} else {
		graph.WriteString(fmt.Sprintf("SPEED: %.2f\n", 0.0))
	}
	graph.WriteString("\n")
	graph.WriteString(fmt.Sprintf("AVERAGE MAGNITUDE: %.2f \n", m.MotionModel.averageMagnitude))
	graph.WriteString(fmt.Sprintf("Average X: %.2f \n", m.MotionModel.xAvg))
	graph.WriteString(fmt.Sprintf("Average Y: %.2f \n", m.MotionModel.yAvg))

	return graph.String()
}

func createAxisGString(gValue float64, axis string) string {
	val := sanitizeGValue(gValue)

	var sb strings.Builder
	sb.WriteString(" | ")

	numberOfDashes := int(math.Abs(val) * 50)

	if val >= 0.0 {
		sb.WriteString(strings.Repeat(" ", 50))
		sb.WriteString("|")
		str := fmt.Sprintf("%s", strings.Repeat(">", numberOfDashes))
		sb.WriteString(str)
		if numberOfDashes < 50 {
			sb.WriteString(fmt.Sprintf("%s", strings.Repeat(" ", 50-numberOfDashes)))
		}
	} else if val < 0.0 {
		if numberOfDashes < 50 {
			sb.WriteString(fmt.Sprintf("%s", strings.Repeat(" ", 50-numberOfDashes)))
		}
		str := fmt.Sprintf("%s", strings.Repeat("<", numberOfDashes))
		sb.WriteString(str)
		sb.WriteString("|")
		sb.WriteString(strings.Repeat(" ", 50))
	}

	sb.WriteString(fmt.Sprintf(" | %s => %.2f \n", axis, val))
	return sb.String()
}

// G force for a car movement can be more and less than 1.0,
// but that is a very rare case and would most often fall between
// 1.0 and -1.0.
func sanitizeGValue(gValue float64) float64 {
	if gValue >= 1.0 {
		return 1.0
	} else if gValue <= -1.0 {
		return -1.0
	} else {
		return gValue
	}
}
