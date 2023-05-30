package tui

import "math"

var magnitudeEvents []float64
var magnitudeSum float64
var magnitudeAverage float64

const (
	TurnThreshold = 0.3 // Threshold value for detecting a turn
)

func computeSpeed(timeInSeconds float64, gForce float64) float64 {
	// Convert g-force to meters per second squared
	acceleration := gForce * 9.8

	// Calculate speed in meters per second
	speed := acceleration * timeInSeconds

	// Convert speed from meters per second to kilometers per hour
	speedKMH := speed * 3.6

	return speedKMH
}

func averageMagnitudeForce(gForceX, gForceY float64) float64 {
	// Calculate the total g-force magnitude
	gForceMagnitude := math.Sqrt(gForceX*gForceX + gForceY*gForceY)
	magnitudeEvents = append(magnitudeEvents, gForceMagnitude)
	magnitudeSum += gForceMagnitude

	if len(magnitudeEvents) == 101 {
		var first float64
		first, magnitudeEvents = magnitudeEvents[0], magnitudeEvents[1:]

		magnitudeSum -= first
		magnitudeAverage = magnitudeSum / 100
	}

	return magnitudeAverage
}
