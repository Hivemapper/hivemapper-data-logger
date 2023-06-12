package imu

import (
	"math"
)

func computeSpeedVariation(timeInSeconds float64, gForce float64) float64 {
	// Convert g-force to meters per second squared
	acceleration := gForce * 9.8

	// Calculate speedVariation in meters per second
	speed := acceleration * timeInSeconds

	// Convert speedVariation from meters per second to kilometers per hour
	speedKMH := speed * 3.6

	return speedKMH
}

func computeTotalMagnitude(xAcceleration float64, yAcceleration float64) float64 {
	return math.Sqrt(math.Pow(xAcceleration, 2) + math.Pow(yAcceleration, 2))
}

// computeTiltAngles Compute the tilt angles ONLY when the total magnitude is of 1.0
func computeTiltAngles(xAcceleration, yAcceleration, zAcceleration float64) (xAngle float64, yAngle float64, zAngle float64) { // returns x, y, z angles
	// http://www.starlino.com/imu_guide.html
	rz := zAcceleration * zAcceleration
	zAngle = 90 * rz
	xAngle = (xAcceleration * xAcceleration) * 90
	yAngle = (yAcceleration * yAcceleration) * 90
	return xAngle, yAngle, zAngle
}

func computeCorrectedGForce(xAcceleration float64, yAcceleration float64, zAcceleration float64, xTilt float64, yTilt float64, zTilt float64) (float64, float64, float64) {
	magnitude := math.Sqrt(xAcceleration*xAcceleration + yAcceleration*yAcceleration + zAcceleration*zAcceleration)

	tiltXacc := math.Sqrt(xTilt / 90)
	tiltYacc := math.Sqrt(yTilt / 90)
	tiltZacc := math.Sqrt(zTilt / 90)

	normalizedX := xAcceleration / magnitude
	normalizedY := yAcceleration / magnitude
	normalizedZ := zAcceleration / magnitude

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

	return correctedGForceX, correctedGForceY, correctedGForceZ
}
