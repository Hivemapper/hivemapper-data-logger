package tui

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/streamingfast/hm-imu-logger/device/iim42652"
)

const GraphHeader = " |                                                                                                       | \n"
const GraphFooter = "-1                                                   0                                                   1 \n"

type Model struct {
	Acceleration *iim42652.Acceleration
	Speed        float64
}

func InitialModel(acceleration *iim42652.Acceleration, speed float64) Model {
	return Model{Acceleration: acceleration, Speed: speed}
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
	case *iim42652.Acceleration:
		m.Acceleration = msg
	case float64:
		m.Speed = msg
	}

	return m, nil
}

func (m Model) View() string {
	var graph strings.Builder
	graph.WriteString(GraphHeader)

	var graphBody strings.Builder

	graphBody.WriteString(createAxisGString(m.Acceleration.CamX(), "X"))
	graphBody.WriteString(createAxisGString(m.Acceleration.CamY(), "Y"))
	graphBody.WriteString(createAxisGString(m.Acceleration.CamZ()-1, "Z"))

	graph.WriteString(graphBody.String())
	graph.WriteString(GraphFooter)
	graph.WriteString(fmt.Sprintf("Speed: %.2f\n", m.Speed))

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
