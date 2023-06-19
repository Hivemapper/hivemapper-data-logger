package imu

import (
	"math"
)

func ComputeSpeedVariation(timeInSeconds float64, gForce float64) float64 {
	// Convert g-force to meters per second squared
	acceleration := gForce * 9.8

	// Calculate speedVariation in meters per second
	speed := acceleration * timeInSeconds

	// Convert speedVariation from meters per second to kilometers per hour
	speedKMH := speed * 3.6

	return speedKMH
}

func ComputeTotalMagnitude(xAcceleration float64, yAcceleration float64) float64 {
	return math.Sqrt(math.Pow(xAcceleration, 2) + math.Pow(yAcceleration, 2))
}

// computeTiltAngles Compute the tilt angles ONLY when the total magnitude is of 1.0
func computeTiltAngles(acceleration *Acceleration) (xAngle float64, yAngle float64, zAngle float64) { // returns x, y, z angles
	// http://www.starlino.com/imu_guide.html
	xAngle = (acceleration.X * acceleration.X) * 90
	yAngle = (acceleration.Y * acceleration.Y) * 90
	zAngle = (acceleration.Z * acceleration.Z) * 90
	return
}

func computeCorrectedGForce(acceleration *Acceleration, xTilt float64, yTilt float64, zTilt float64) *Acceleration {
	x := acceleration.X
	y := acceleration.Y
	z := acceleration.Z
	m := acceleration.Magnitude

	tiltXacc := math.Sqrt(xTilt / 90)
	tiltYacc := math.Sqrt(yTilt / 90)
	tiltZacc := math.Sqrt(zTilt / 90)

	normalizedX := x / m
	normalizedY := y / m
	normalizedZ := z / m

	normalizedXY := normalizedX + normalizedY
	distRatioX := 0.0
	distRatioY := 0.0
	if normalizedXY != 0 {
		distRatioX = normalizedX / normalizedXY
		distRatioY = normalizedY / normalizedXY
	}

	zDelta := 1 - tiltZacc
	correctedGForceZ := tiltZacc + zDelta
	foo := normalizedZ - tiltZacc

	correctedGForceX := normalizedX - tiltXacc - (foo * distRatioX)
	correctedGForceY := normalizedY - tiltYacc - (foo * distRatioY)

	//fmt.Println("correctedGForceX", correctedGForceX, "correctedGForceY", correctedGForceY, "correctedGForceZ", correctedGForceZ)
	return NewAcceleration(correctedGForceX, correctedGForceY, correctedGForceZ, m, acceleration.Temperature, acceleration.Time)
}
