package imu

import (
	"github.com/streamingfast/imu-controller/device/iim42652"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_CameraMountOrientation(t *testing.T) {
	tests := []struct {
		name                string
		rawEvent            *RawImuEvent
		expectedOrientation Orientation
	}{
		{
			name: "front camera orientation",
			rawEvent: &RawImuEvent{
				Acceleration: newAcceleration(0.5, 0.0, 1.0),
			},
			expectedOrientation: OrientationFront,
		},
		{
			name: "right side camera orientation",
			rawEvent: &RawImuEvent{
				Acceleration: newAcceleration(0.0, -0.5, 1.0),
			},
			expectedOrientation: OrientationRight,
		},
		{
			name: "left side camera orientation",
			rawEvent: &RawImuEvent{
				Acceleration: newAcceleration(0.0, 0.5, 1.0),
			},
			expectedOrientation: OrientationLeft,
		},
		{
			name: "back camera orientation",
			rawEvent: &RawImuEvent{
				Acceleration: newAcceleration(-0.5, 0.0, 1.0),
			},
			expectedOrientation: OrientationBack,
		},
		{
			name: "don't know for sure the position of the camera",
			rawEvent: &RawImuEvent{
				Acceleration: newAcceleration(0.5, 0.5, 1.0),
			},
			expectedOrientation: OrientationUnset,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expectedOrientation, computeOrientation(test.rawEvent))
		})
	}
}

func newAcceleration(x, y, z float64) *iim42652.Acceleration {
	// Z -> CamX()
	// X -> CamY()
	// Y -> CamY()
	return &iim42652.Acceleration{
		Z: x, X: y, Y: z,
	}
}
