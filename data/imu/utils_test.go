package imu

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ComputeAccelerationSpeed(t *testing.T) {
	tests := []struct {
		name          string
		timeInSeconds float64
		gForce        float64
		expectedSpeed float64
	}{
		{
			name:          "stopped car",
			timeInSeconds: 0.0,
			gForce:        0.0,
			expectedSpeed: 0.0,
		},
		{
			name:          "normally expected 1.0g 0-60 mph acceleration",
			timeInSeconds: 2.8,
			gForce:        1.0,
			expectedSpeed: 98.784,
		},
		{
			name:          "average deceleration 0.30g over 5s",
			timeInSeconds: 5,
			gForce:        -0.30,
			expectedSpeed: -52.92,
		},
		{
			name:          "average driver max deceleration 0.47 over 5s",
			timeInSeconds: 5,
			gForce:        -0.47,
			expectedSpeed: -82.908,
		},
		{
			name:          "vehicle max deceleration 0.70 over 5s",
			timeInSeconds: 5,
			gForce:        -0.70,
			expectedSpeed: -123.48000000000002,
		},
		{
			name:          "normally expected 1.0g deceleration",
			timeInSeconds: 5,
			gForce:        -1.0,
			expectedSpeed: -176.4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expectedSpeed, ComputeSpeedVariation(test.timeInSeconds, test.gForce))
		})
	}
}

func Test_ComputeCorrectedGForce(t *testing.T) {
	tests := []struct {
		name           string
		xAcceleration  float64
		yAcceleration  float64
		zAcceleration  float64
		xAngle         float64
		yAngle         float64
		zAngle         float64
		expectedXValue float64
		expectedYValue float64
		expectedZValue float64
	}{
		{
			name:           "45 degree tilt",
			xAcceleration:  0.707106781186548,
			yAcceleration:  0.0,
			zAcceleration:  0.707106781186548,
			xAngle:         45.0,
			yAngle:         0.0,
			zAngle:         45.0,
			expectedXValue: 0.0,
			expectedYValue: 0.0,
			expectedZValue: 1.0,
		},
		{
			name:           "45 degree tilts + 0.1g acceleration",
			xAcceleration:  0.807106781186548,
			yAcceleration:  0.0,
			zAcceleration:  0.707106781186548,
			xAngle:         45.0,
			yAngle:         0.0,
			zAngle:         45.0,
			expectedXValue: 0.093,
			expectedYValue: 0.0,
			expectedZValue: 1.0,
		},
		{
			name:           "45 degree tilts + 0.1g deceleration",
			xAcceleration:  0.607106781186548,
			yAcceleration:  0.0,
			zAcceleration:  0.707106781186548,
			xAngle:         45.0,
			yAngle:         0.0,
			zAngle:         45.0,
			expectedXValue: -0.107,
			expectedYValue: 0.0,
			expectedZValue: 1.0,
		},
		{
			name:           "Flat",
			xAcceleration:  0.0,
			yAcceleration:  0.0,
			zAcceleration:  1.0,
			xAngle:         0.0,
			yAngle:         0.0,
			zAngle:         90.0,
			expectedXValue: 0.0,
			expectedYValue: 0.0,
			expectedZValue: 1.0,
		},
		{
			name:           "Flat + 0.1g acceleration",
			xAcceleration:  0.1,
			yAcceleration:  0.0,
			zAcceleration:  1.0,
			xAngle:         0.0,
			yAngle:         0.0,
			zAngle:         90.0,
			expectedXValue: 0.104,
			expectedYValue: 0.0,
			expectedZValue: 1.0,
		},
		{
			name:           "Flat + 0.1g acceleration in curve",
			xAcceleration:  0.25,
			yAcceleration:  0.25,
			zAcceleration:  1.0,
			xAngle:         0.0,
			yAngle:         0.0,
			zAngle:         90.0,
			expectedXValue: 0.264,
			expectedYValue: 0.264,
			expectedZValue: 1.0,
		},
		{
			name:           "Flat + 0.1g deceleration",
			xAcceleration:  -0.1,
			yAcceleration:  0.0,
			zAcceleration:  1.0,
			xAngle:         0.0,
			yAngle:         0.0,
			zAngle:         90.0,
			expectedXValue: -0.095,
			expectedYValue: 0.0,
			expectedZValue: 1.0,
		},
		{
			name:           "Flat + 0.1g deceleration and going right",
			xAcceleration:  -0.1,
			yAcceleration:  0.1,
			zAcceleration:  1.0,
			xAngle:         0.0,
			yAngle:         0.0,
			zAngle:         90.0,
			expectedXValue: -0.099,
			expectedYValue: 0.099,
			expectedZValue: 1.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			correctedX, correctedY, correctedZ := computeCorrectedGForce(test.xAcceleration, test.yAcceleration, test.zAcceleration, test.xAngle, test.yAngle, test.zAngle)
			r := func(v float64) float64 {
				if v == 0 {
					return v
				}
				return math.Round(v*1000) / 1000
			}
			correctedX = r(correctedX)
			correctedY = r(correctedY)
			fmt.Printf("%f\n", correctedX)
			fmt.Printf("%f\n", correctedY)
			require.Equal(t, test.expectedXValue, correctedX)
			require.Equal(t, test.expectedYValue, correctedY)
			require.Equal(t, test.expectedZValue, correctedZ)

		})
	}
}

func Test_ComputeTiltAngles(t *testing.T) {
	tests := []struct {
		name                string
		xAxis               float64
		yAxis               float64
		zAxis               float64
		expectedXAngleValue float64
		expectedYAngleValue float64
	}{
		{
			name:                "flat",
			xAxis:               0.0,
			yAxis:               0.0,
			zAxis:               1.0,
			expectedXAngleValue: 0.0,
			expectedYAngleValue: 0.0,
		},
		{
			name:                "",
			xAxis:               0.707106781186548,
			yAxis:               0.0,
			zAxis:               0.707106781186548,
			expectedXAngleValue: 45.0,
			expectedYAngleValue: 0.0,
		},
		{
			name:                "",
			xAxis:               -0.707106781186548,
			yAxis:               0.0,
			zAxis:               0.707106781186548,
			expectedXAngleValue: -45.0,
			expectedYAngleValue: 0.0,
		},
		{
			name:                "acceleration, no turn",
			xAxis:               0.1,
			yAxis:               0.0,
			zAxis:               1.0,
			expectedXAngleValue: 5.710593137499643,
			expectedYAngleValue: 0.0,
		},
		{
			name:                "deceleration, no turn",
			xAxis:               -0.1,
			yAxis:               0.0,
			zAxis:               1.0,
			expectedXAngleValue: -5.710593137499643,
			expectedYAngleValue: 0.0,
		},
		{
			name:                "acceleration, right turn",
			xAxis:               0.1,
			yAxis:               0.1,
			zAxis:               1.0,
			expectedXAngleValue: 5.6824384835168384,
			expectedYAngleValue: 5.6824384835168384,
		},
		{
			name:                "deceleration, right turn",
			xAxis:               -0.1,
			yAxis:               0.1,
			zAxis:               1.0,
			expectedXAngleValue: -5.6824384835168384,
			expectedYAngleValue: 5.6824384835168384,
		},
		{
			name:                "acceleration, left turn",
			xAxis:               0.1,
			yAxis:               -0.1,
			zAxis:               1.0,
			expectedXAngleValue: 5.6824384835168384,
			expectedYAngleValue: -5.6824384835168384,
		},
		{
			name:                "deceleration, left turn",
			xAxis:               -0.1,
			yAxis:               -0.1,
			zAxis:               1.0,
			expectedXAngleValue: -5.6824384835168384,
			expectedYAngleValue: -5.6824384835168384,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			x, y, _ := computeTiltAngles(test.xAxis, test.yAxis, test.zAxis)

			require.Equal(t, test.expectedXAngleValue, x)
			require.Equal(t, test.expectedYAngleValue, y)
		})
	}
}
