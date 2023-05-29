package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

const GraphHeader = " |                                                                                                       | \n"
const GraphFooter = "-1                                                   0                                                   1 \n"

type Model struct {
	XAxisG float64
	YAxisG float64
	ZAxisG float64
}

func initialModel() Model {
	return Model{
		XAxisG: 0.0,
		YAxisG: 0.0,
		ZAxisG: 0.0,
	}
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
	}

	return m, nil
}

func (m Model) View() string {
	var graph strings.Builder
	graph.WriteString(GraphHeader)

	var graphBody strings.Builder

	graphBody.WriteString(createAxisGString(m.XAxisG, "X"))
	graphBody.WriteString(createAxisGString(m.YAxisG, "Y"))
	graphBody.WriteString(createAxisGString(m.ZAxisG, "Z"))

	graph.WriteString(graphBody.String())
	graph.WriteString(GraphFooter)

	return graph.String()
}

func createAxisGString(gValue float64, axis string) string {
	val := sanitizeGValue(gValue)

	var sb strings.Builder
	sb.WriteString(" | ")

	str := fmt.Sprintf("%s", strings.Repeat("-", int(val*50)))

	if val >= 1.0 {
		sb.WriteString(strings.Repeat(" ", 50))
		sb.WriteString("-")
		sb.WriteString(str)
	} else if val < 1.0 {
		sb.WriteString(str)
		sb.WriteString("-")
		sb.WriteString(strings.Repeat(" ", 50))
	}

	sb.WriteString(fmt.Sprintf(" | %s => %.2f", axis, val))
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
