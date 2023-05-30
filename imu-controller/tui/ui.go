package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"imu-logger/device/iim42652"
	"math"
	"strings"
)

const GraphHeader = " |                                                                                                       | \n"
const GraphFooter = "-1                                                   0                                                   1 \n"

type Model struct {
	Acceleration *iim42652.Acceleration
}

func InitialModel(acceleration *iim42652.Acceleration) Model {
	return Model{Acceleration: acceleration}
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
	}

	return m, nil
}

func (m Model) View() string {
	var graph strings.Builder
	graph.WriteString(GraphHeader)

	var graphBody strings.Builder

	graphBody.WriteString(createAxisGString(m.Acceleration.X, "X"))
	graphBody.WriteString(createAxisGString(m.Acceleration.Y, "Y"))
	graphBody.WriteString(createAxisGString(m.Acceleration.Z, "Z"))

	graph.WriteString(graphBody.String())
	graph.WriteString(GraphFooter)

	return graph.String()
}

func createAxisGString(gValue float64, axis string) string {
	val := sanitizeGValue(gValue)

	var sb strings.Builder
	sb.WriteString(" | ")

	numberOfDashes := int(math.Abs(val) * 50)
	str := fmt.Sprintf("%s", strings.Repeat("-", numberOfDashes))

	if val >= 1.0 {
		sb.WriteString(strings.Repeat(" ", 50))
		sb.WriteString("-")
		sb.WriteString(str)
		if numberOfDashes < 50 {
			sb.WriteString(fmt.Sprintf("%s", strings.Repeat(" ", 50-numberOfDashes)))
		}
	} else if val < 1.0 {
		if numberOfDashes < 50 {
			sb.WriteString(fmt.Sprintf("%s", strings.Repeat(" ", 50-numberOfDashes)))
		}
		sb.WriteString(str)
		sb.WriteString("-")
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
