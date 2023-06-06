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

func computeCorrectedTiltAngles(gForceX, gForceY, gForceZ float64) (float64, float64, float64) { // returns x, y, z angles
	// http://www.starlino.com/imu_guide.html
	//We can deduct from Eq.1 that R = SQRT( Rx^2 + Ry^2 + Rz^2).
	//
	//	We can find now our angles by using arccos() function (the inverse cos() function ):
	// Axr = arccos(Rx/R)
	//Ayr = arccos(Ry/R)
	//Azr = arccos(Rz/R)

	// https://engineering.stackexchange.com/questions/3348/calculating-pitch-yaw-and-roll-from-mag-acc-and-gyro-data
	accelerationX := gForceX * 3.9
	accelerationY := gForceY * 3.9
	accelerationZ := gForceZ * 3.9
	pitchX := 180 * math.Atan(accelerationX/math.Sqrt(accelerationY*accelerationY+accelerationZ*accelerationZ)) / math.Pi
	rollY := 180 * math.Atan(accelerationY/math.Sqrt(accelerationX*accelerationX+accelerationZ*accelerationZ)) / math.Pi
	yawZ := 180 * math.Atan(accelerationZ/math.Sqrt(accelerationX*accelerationX+accelerationY*accelerationY)) / math.Pi

	fmt.Println("stackoverflow pitch", pitchX)
	fmt.Println("stackoverflow roll", rollY)
	fmt.Println("stackoverflow yaw", yawZ)

	return pitchX, rollY, yawZ
}

func computeCorrectedGForceAxis(zAxis, complementaryAxis float64) float64 {
	tiltAngleXRad := math.Atan2(zAxis, complementaryAxis)
	correctedValue := complementaryAxis * math.Cos(tiltAngleXRad)

	return correctedValue
}

func computeCorrectedGForce(xAcceleration, yAcceleration, zAcceleration, xAngle, yAngle, zAngle float64) (float64, float64, float64) {
	//// Compute x force with no acceleration. The x value is the computation at the exact angle.
	//noAccelerationX := math.Sqrt(xAngle / 90)
	//correctedX := xAcceleration - noAccelerationX
	//
	//// Compute x force with no acceleration. The x value is the computation at the exact angle.
	//noAccelerationY := math.Sqrt(yAngle / 90)
	//correctedY := yAcceleration - noAccelerationY
	//
	//// We want this value to be 1.0 for the moment
	//correctedZ := 1.0

	magnitude := math.Sqrt(xAcceleration*xAcceleration + yAcceleration*yAcceleration + zAcceleration*zAcceleration)
	normalizedX := xAcceleration / magnitude
	normalizedY := yAcceleration / magnitude
	normalizedZ := zAcceleration / magnitude

	angleXRad := xAngle * math.Pi / 180.0
	angleYRad := yAngle * math.Pi / 180.0
	angleZRad := zAngle * math.Pi / 180.0

	cosX := math.Cos(angleXRad)
	cosY := math.Cos(angleYRad)
	cosZ := math.Cos(angleZRad)

	correctedX := normalizedX / cosY / cosZ
	correctedY := normalizedY / cosX / cosZ
	correctedZ := normalizedZ / cosX / cosY

	//correctedX := xAcceleration * math.Cos(angleXRad)
	//correctedY := yAcceleration * math.Cos(angleYRad)
	//correctedZ := zAcceleration * math.Cos(angleZRad)

	return correctedX, correctedY, correctedZ
}
