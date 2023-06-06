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
		name              string
		zAxis             float64
		complimentaryAxis float64
		expectedValue     float64
	}{
		{
			name:              "flat on table",
			zAxis:             1.0,
			complimentaryAxis: 0.0,
			expectedValue:     0.0,
		},
		{
			name:              "modelled data",
			zAxis:             0.9999999952985156,
			complimentaryAxis: 0.0028501808113100013,
			expectedValue:     0.0,
		},
		{
			name:              "data",
			zAxis:             1.01,
			complimentaryAxis: 0.01,
			expectedValue:     0.0,
		}, {
			name:              "2", //0.42205761869463915 -0.005083033799155377 0.9014656664890688
			zAxis:             0.9014656664890688,
			complimentaryAxis: 0.42205761869463915,
			expectedValue:     0.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expectedValue, computeCorrectedGForce(test.zAxis, test.complimentaryAxis))
		})
	}
}

//func Test_ComputeSpeed(t *testing.T) {
//	tests := []struct {
//		name               string
//		accelerationSpeeds []float64
//		expectedSpeed      float64
//	}{
//		{
//			name:               "stopped car",
//			accelerationSpeeds: []float64{100.0, -50.0, 10.0},
//			expectedSpeed:      60.0,
//		},
//	}
//
//	for _, test := range tests {
//		t.Start(test.name, func(t *testing.T) {
//			for _, accelSpeed := range test.accelerationSpeeds {
//				addAccelerationSpeeds(accelSpeed)
//			}
//			require.Equal(t, test.expectedSpeed, computeSpeed())
//		})
//	}
//}
