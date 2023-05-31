package tui

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

const GraphHeader = " |                                                                                                       | \n"
const GraphFooter = "-1                                                   0                                                   1 \n"

type MotionModel struct {
	Acceleration     *iim42652.Acceleration
	speedVariation   float64
	speed            float64
	averageMagnitude float64
	xAvg             float64
	yAvg             float64
	events           []string
	gnssData         *neom9n.Data
}

type Model struct {
	MotionModel *MotionModel
}

type MotionModelMsg struct {
	Acceleration   *iim42652.Acceleration
	speedVariation float64
	speed          float64
	xAvg           *data.AverageFloat64
	yAvg           *data.AverageFloat64
	magnitudeAvg   *data.AverageFloat64
	event          string
	gnssData       *neom9n.Data
}

func NewModel() *Model {
	return &Model{
		MotionModel: &MotionModel{
			Acceleration:     &iim42652.Acceleration{},
			speed:            0.0,
			averageMagnitude: 0.0,
			gnssData: &neom9n.Data{
				Dop: &neom9n.Dop{},
				RF:  &neom9n.RF{},
			},
		},
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
	case *MotionModelMsg:
		if msg.Acceleration != nil {
			m.MotionModel.Acceleration = msg.Acceleration
		}
		if msg.xAvg != nil {
			m.MotionModel.xAvg = msg.xAvg.Average
		}
		if msg.yAvg != nil {
			m.MotionModel.yAvg = msg.yAvg.Average
		}
		if msg.magnitudeAvg != nil {
			m.MotionModel.averageMagnitude = msg.magnitudeAvg.Average
		}

		m.MotionModel.speedVariation = msg.speedVariation
		m.MotionModel.speed = msg.speed
		if msg.event != "" {
			m.MotionModel.events = append([]string{msg.event}, m.MotionModel.events...)
			if len(m.MotionModel.events) > 10 {
				m.MotionModel.events = m.MotionModel.events[:len(m.MotionModel.events)-1]
			}
		}

		if msg.gnssData != nil {
			m.MotionModel.gnssData = msg.gnssData
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

	graph.WriteString("\nGPS DATA\n")
	graph.WriteString(createGnssDataGString(m.MotionModel.gnssData))
	graph.WriteString("\n")

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

	var eventsSb strings.Builder
	for _, event := range m.MotionModel.events {
		eventsSb.WriteString(fmt.Sprintf("\t%s \n", event))
	}

	graph.WriteString("Events:\n")
	graph.WriteString(eventsSb.String())

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

func createGnssDataGString(gnssData *neom9n.Data) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\t ttff: %d Heading: %f Speed: %f\n", gnssData.Ttff, gnssData.Heading, gnssData.Speed))
	sb.WriteString(fmt.Sprintf("\t Latitude: %f Longitude: %f\n", gnssData.Latitude, gnssData.Longitude))
	sb.WriteString(fmt.Sprintf("\t HDop: %f VDop: %f Eph: %f \n", gnssData.Dop.HDop, gnssData.Dop.VDop, gnssData.Eph))
	sb.WriteString("\n")
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
