package imu

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_CameraMountOrientation(t *testing.T) {
	tests := []struct {
		name                string
		acceleration        *Acceleration
		expectedOrientation Orientation
	}{
		{
			name:                "front camera orientation",
			acceleration:        NewAcceleration(0.5, 0.0, 1.0, -99, time.Now().UTC()),
			expectedOrientation: OrientationFront,
		},
		{
			name:                "right side camera orientation",
			acceleration:        NewAcceleration(0.0, 0.5, 1.0, -99, time.Now().UTC()),
			expectedOrientation: OrientationRight,
		},
		{
			name:                "left side camera orientation",
			acceleration:        NewAcceleration(0.0, -0.5, 1.0, -99, time.Now().UTC()),
			expectedOrientation: OrientationLeft,
		},
		{
			name:                "back camera orientation",
			acceleration:        NewAcceleration(-0.5, 0.0, 1.0, -99, time.Now().UTC()),
			expectedOrientation: OrientationBack,
		},
		{
			name:                "a bit of front acceleration, but not enough to know the orientation",
			acceleration:        NewAcceleration(0.05, 0.0, 1.0, -99, time.Now().UTC()),
			expectedOrientation: OrientationUnset,
		},
		{
			name:                "a bit of back acceleration, but not enough to know the orientation",
			acceleration:        NewAcceleration(-0.05, 0.0, 1.0, -99, time.Now().UTC()),
			expectedOrientation: OrientationUnset,
		},
		{
			name:                "a bit of right side acceleration, but not enough to know the orientation",
			acceleration:        NewAcceleration(0.0, 0.05, 1.0, -99, time.Now().UTC()),
			expectedOrientation: OrientationUnset,
		},
		{
			name:                "a bit of left side acceleration, but not enough to know the orientation",
			acceleration:        NewAcceleration(0.0, -0.05, 1.0, -99, time.Now().UTC()),
			expectedOrientation: OrientationUnset,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expectedOrientation, computeOrientation(test.acceleration))
		})
	}
}
