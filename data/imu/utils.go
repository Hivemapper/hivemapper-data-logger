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

// todo: do we need to compute the euler angles instead ?
func computeTiltAngle(zAxis, complementaryAxis float64) float64 {
	tiltAngleXRad := math.Atan2(zAxis, complementaryAxis)
	tiltAngleXDegrees := tiltAngleXRad * (180 / math.Pi)

	return tiltAngleXDegrees
}

func computeTiltAngles(xAcceleration, yAcceleration, zAcceleration float64) (xAngle float64, yAngle float64, zAngle float64) { // returns x, y, z angles
	// http://www.starlino.com/imu_guide.html
	//We can deduct from Eq.1 that R = SQRT( Rx^2 + Ry^2 + Rz^2).
	//
	//	We can find now our angles by using arccos() function (the inverse cos() function ):
	// Axr = arccos(Rx/R)
	//Ayr = arccos(Ry/R)
	//Azr = arccos(Rz/R)

	// https://engineering.stackexchange.com/questions/3348/calculating-pitch-yaw-and-roll-from-mag-acc-and-gyro-data
	//accelerationX := gForceX * 3.9
	//accelerationY := gForceY * 3.9
	//accelerationZ := gForceZ * 3.9
	//accelerationX := gForceX
	//accelerationY := gForceY
	//accelerationZ := gForceZ
	//xAngle = 180 * math.Atan(accelerationX/math.Sqrt(accelerationY*accelerationY+accelerationZ*accelerationZ)) / math.Pi
	//yAngle = 180 * math.Atan(accelerationY/math.Sqrt(accelerationX*accelerationX+accelerationZ*accelerationZ)) / math.Pi
	magnitude := math.Sqrt(xAcceleration*xAcceleration + yAcceleration*yAcceleration + zAcceleration*zAcceleration)
	xAngle = math.Acos(xAcceleration/magnitude) * 180 / math.Pi
	yAngle = math.Acos(yAcceleration/magnitude) * 180 / math.Pi
	zAngle = math.Acos(yAcceleration/magnitude) * 180 / math.Pi
	return xAngle, yAngle, zAngle
}

func computeCorrectedGForceAxis(zAxis, complementaryAxis float64) float64 {
	tiltAngleXRad := math.Atan2(zAxis, complementaryAxis)
	correctedValue := complementaryAxis * math.Cos(tiltAngleXRad)

	return correctedValue
}

func computeCorrectedGForce(xAcceleration float64, yAcceleration float64, zAcceleration float64) (float64, float64) {
	rz := zAcceleration * zAcceleration
	zAngle := 90 * rz

	xBaseAcceleration := math.Cos(zAngle * math.Pi / 180)

	var x float64
	if xBaseAcceleration > 0 {
		x = xAcceleration - xBaseAcceleration
	} else {
		x = xAcceleration + xBaseAcceleration
	}

	return x, yAcceleration
}
