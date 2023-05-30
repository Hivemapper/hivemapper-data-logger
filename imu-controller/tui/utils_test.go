package tui

import (
	"github.com/stretchr/testify/require"
	"testing"
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
			require.Equal(t, test.expectedSpeed, computeAccelerationSpeed(test.timeInSeconds, test.gForce))
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
//		t.Run(test.name, func(t *testing.T) {
//			for _, accelSpeed := range test.accelerationSpeeds {
//				addAccelerationSpeeds(accelSpeed)
//			}
//			require.Equal(t, test.expectedSpeed, computeSpeed())
//		})
//	}
//}
