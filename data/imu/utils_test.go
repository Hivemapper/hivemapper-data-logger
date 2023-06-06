package imu

import (
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
		xAngles        float64
		yAngles        float64
		zAngles        float64
		xAcceleration  float64
		yAcceleration  float64
		zAcceleration  float64
		expectedXValue float64
		expectedYValue float64
		expectedZValue float64
	}{
		//*imu.RawImuAccelerationEvent Event: RawImuAccelerationEvent Acceleration{camX:0.57570, camY:0.00293, camZ: 0.80764, totalMagn: 0.99183}
		//*imu.CorrectedAccelerationEvent Event: CorrectedAccelerationEvent: -0.052186, -0.040435, 1.000000 Angles x 35.481755, y 0.169247, z 54.518245
		{
			name:           "45 degree tilt",
			xAngles:        45.0,
			yAngles:        0.0,
			zAngles:        45.0,
			xAcceleration:  0.71,
			yAcceleration:  0.00,
			zAcceleration:  0.71,
			expectedXValue: 0.0,
			expectedYValue: 0.0,
			expectedZValue: 1.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			correctedX, correctedY, correctedZ := computeCorrectedGForce(test.xAcceleration, test.yAcceleration, test.zAcceleration, test.xAngles, test.yAngles, test.zAngles)
			require.Equal(t, test.expectedXValue, correctedX)
			require.Equal(t, test.expectedYValue, correctedY)
			require.Equal(t, test.expectedZValue, correctedZ)
		})
	}
}

func Test_ComputeCorrectedTiltAngles(t *testing.T) {
	tests := []struct {
		name                string
		xAxis               float64
		yAxis               float64
		zAxis               float64
		expectedXAngleValue float64
		expectedYAngleValue float64
		expectedZAngleValue float64
	}{
		{
			name:                "no acceleration, driving on a 100% plain field",
			xAxis:               0.0,
			yAxis:               0.0,
			zAxis:               1.0,
			expectedXAngleValue: 0.0,
			expectedYAngleValue: 0.0,
			expectedZAngleValue: 90.0,
		},
		{
			name:                "acceleration, no turn",
			xAxis:               0.1,
			yAxis:               0.0,
			zAxis:               1.0,
			expectedXAngleValue: 5.710593137499643,
			expectedYAngleValue: 0.0,
			expectedZAngleValue: 95.710593137499643,
		},
		{
			name:                "deceleration, no turn",
			xAxis:               -0.1,
			yAxis:               0.0,
			zAxis:               1.0,
			expectedXAngleValue: -5.710593137499643,
			expectedYAngleValue: 0.0,
			expectedZAngleValue: 84.28940686250036,
		},
		{
			name:                "acceleration, right turn",
			xAxis:               0.1,
			yAxis:               0.1,
			zAxis:               1.0,
			expectedXAngleValue: 5.6824384835168384,
			expectedYAngleValue: 5.6824384835168384,
			expectedZAngleValue: 95.68243848351683,
		},
		{
			name:                "deceleration, right turn",
			xAxis:               -0.1,
			yAxis:               0.1,
			zAxis:               1.0,
			expectedXAngleValue: -5.6824384835168384,
			expectedYAngleValue: 5.6824384835168384,
			expectedZAngleValue: 84.31756151648317,
		},
		{
			name:                "acceleration, left turn",
			xAxis:               0.1,
			yAxis:               -0.1,
			zAxis:               1.0,
			expectedXAngleValue: 5.6824384835168384,
			expectedYAngleValue: -5.6824384835168384,
			expectedZAngleValue: 95.68243848351683,
		},
		{
			name:                "deceleration, left turn",
			xAxis:               -0.1,
			yAxis:               -0.1,
			zAxis:               1.0,
			expectedXAngleValue: -5.6824384835168384,
			expectedYAngleValue: -5.6824384835168384,
			expectedZAngleValue: 84.31756151648317,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			correctedX, correctedY, correctedZ := computeCorrectedTiltAngles(test.xAxis, test.yAxis, test.zAxis)
			require.Equal(t, test.expectedXAngleValue, correctedX)
			require.Equal(t, test.expectedYAngleValue, correctedY)
			require.Equal(t, test.expectedZAngleValue, correctedZ)
		})
	}
}
