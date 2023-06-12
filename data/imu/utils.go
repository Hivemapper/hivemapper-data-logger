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

func computeCorrectedGForceAxis(zAxis, complementaryAxis float64) float64 {
	tiltAngleXRad := math.Atan2(zAxis, complementaryAxis)
	correctedValue := complementaryAxis * math.Cos(tiltAngleXRad)

	return correctedValue
}

func computeCorrectedGForce(xAcceleration float64, yAcceleration float64, zAngle float64) (float64, float64) {
	xBaseAcceleration := math.Cos(zAngle * math.Pi / 180)

	var x float64
	if xBaseAcceleration > 0 {
		x = xAcceleration - xBaseAcceleration
	} else {
		x = xAcceleration + xBaseAcceleration
	}

	return x, yAcceleration
}
