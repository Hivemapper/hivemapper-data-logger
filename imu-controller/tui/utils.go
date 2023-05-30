package tui

func computeSpeed(timeInSeconds float64, gForce float64) float64 {
	// Convert g-force to meters per second squared
	acceleration := gForce * 9.8

	// Calculate speed in meters per second
	speed := acceleration * timeInSeconds

	// Convert speed from meters per second to kilometers per hour
	speedKMH := speed * 3.6

	return speedKMH
}
