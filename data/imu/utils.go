package imu

import (
	"fmt"
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

// todo: do we need to compute the euler angles instead ?
func computeTiltAngle(zAxis, complementaryAxis float64) float64 {
	tiltAngleXRad := math.Atan2(zAxis, complementaryAxis)
	tiltAngleXDegrees := tiltAngleXRad * (180 / math.Pi)

	return tiltAngleXDegrees
}

func computeCorrectedGForce(zAxis, complementaryAxis float64) float64 {
	tiltAngleXRad := math.Atan2(zAxis, complementaryAxis)
	tiltAngleXDegrees := tiltAngleXRad * math.Pi
	fmt.Println("angle before", tiltAngleXDegrees)
	tiltAngleXDegrees = 90 - tiltAngleXDegrees

	tiltAngleXRad = tiltAngleXDegrees * math.Pi / 180.0

	fmt.Println("angle", tiltAngleXDegrees)

	correctedValue := complementaryAxis * math.Cos(tiltAngleXRad)

	return correctedValue
}
