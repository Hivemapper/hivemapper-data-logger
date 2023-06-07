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
			require.Equal(t, test.expectedSpeed, computeSpeedVariation(test.timeInSeconds, test.gForce))
		})
	}
}

func Test_ComputeCorrectedGForce(t *testing.T) {
	tests := []struct {
		name           string
		xAcceleration  float64
		yAcceleration  float64
		zAcceleration  float64
		expectedXValue float64
		expectedYValue float64
	}{
		{
			name:           "45 degree tilt",
			xAcceleration:  0.707106781186548,
			yAcceleration:  0.0,
			zAcceleration:  0.707106781186548,
			expectedXValue: 0.0,
			expectedYValue: 0.0,
		},
		{
			name:           "45 degree tilt",
			xAcceleration:  0.0,
			yAcceleration:  0.707106781186548,
			zAcceleration:  0.707106781186548,
			expectedXValue: 0.0,
			expectedYValue: 0.0,
		},
		{
			name:           "flat",
			zAcceleration:  1.0,
			expectedXValue: 0.0,
			expectedYValue: 0.0,
		},
		{
			name:           "flat fast acceleration",
			xAcceleration:  0.890652,
			zAcceleration:  1.009796,
			expectedXValue: 0.23,
			expectedYValue: 0.0,
		},
		{
			name:           "Tilted x and y but not moving",
			xAcceleration:  0.20850,
			yAcceleration:  0.19337,
			zAcceleration:  0.95999,
			expectedXValue: 0.0,
			expectedYValue: 0.0,
		},
		{
			name:           "Tilted x and y but not moving",
			xAcceleration:  0.47267,
			yAcceleration:  0.01660,
			zAcceleration:  0.88626,
			expectedXValue: 0.0,
			expectedYValue: 0.0,
		},
		{
			name:           "flat fast acceleration",
			xAcceleration:  0.30274361400189215,
			yAcceleration:  0.10986663411358989,
			zAcceleration:  0.978057191686758,
			expectedXValue: 0.0,
			expectedYValue: 0.0,
		},
		//0.2568437757499924,0.021485030671102023,1.000030518509476
		//0.23291726432081056,0.02294991912594989,0.9804986724448378
		//0.21338541825617235,0.11670278023621326,1.1807000946073793
		{
			name:           "acceleration",
			xAcceleration:  0.2568437757499924,
			yAcceleration:  0.021485030671102023,
			zAcceleration:  1.000030518509476,
			expectedXValue: 0.0,
			expectedYValue: 0.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			correctedX, correctedY := computeCorrectedGForce(test.xAcceleration, test.yAcceleration, test.zAcceleration)
			r := func(v float64) float64 {
				return math.Round(v*100) / 100
			}
			correctedX = r(correctedX)
			correctedY = r(correctedY)
			fmt.Printf("%f\n", correctedX)
			fmt.Printf("%f\n", correctedY)
			require.Equal(t, test.expectedXValue, correctedX)
			require.Equal(t, test.expectedYValue, correctedY)

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
			x, y := computeTiltAngles(test.xAxis, test.yAxis, test.zAxis)

			require.Equal(t, test.expectedXAngleValue, x)
			require.Equal(t, test.expectedYAngleValue, y)
		})
	}
}
