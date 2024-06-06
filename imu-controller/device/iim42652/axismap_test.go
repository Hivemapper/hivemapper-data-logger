package iim42652

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_AxisMap(t *testing.T) {
	tests := []struct {
		name                  string
		axisMap               *AxisMap
		acceleration          *Acceleration
		expectedAccelerationX float64
		expectedAccelerationY float64
		expectedAccelerationZ float64
	}{
		{
			name: "not inverted and cam x = x axis mapping",
			axisMap: &AxisMap{
				CamX: "X",
				CamY: "Y",
				CamZ: "Z",
				InvX: false,
				InvY: false,
				InvZ: false,
			},
			acceleration: &Acceleration{
				X: 0.1,
				Y: 0.0,
				Z: 0.0,
			},
			expectedAccelerationX: 0.1,
		},
		{
			name: "inverted x and cam x = x axis mapping",
			axisMap: &AxisMap{
				CamX: "X",
				CamY: "Y",
				CamZ: "Z",
				InvX: true,
				InvY: false,
				InvZ: false,
			},
			acceleration: &Acceleration{
				X: 0.1,
			},
			expectedAccelerationX: -0.1,
		},
		{
			name: "inverted x and cam x <> x axis mapping",
			axisMap: &AxisMap{
				CamX: "Y",
				CamY: "Z",
				CamZ: "X",
				InvX: true,
				InvY: false,
				InvZ: false,
			},
			acceleration: &Acceleration{
				Y: 0.1,
			},
			expectedAccelerationX: -0.1,
		},
		{
			name: "not inverted and cam y = y axis mapping",
			axisMap: &AxisMap{
				CamX: "X",
				CamY: "Y",
				CamZ: "Z",
				InvX: false,
				InvY: false,
				InvZ: false,
			},
			acceleration: &Acceleration{
				Y: 0.1,
			},
			expectedAccelerationY: 0.1,
		},
		{
			name: "inverted y and cam y = y axis mapping",
			axisMap: &AxisMap{
				CamX: "X",
				CamY: "Y",
				CamZ: "Z",
				InvX: false,
				InvY: true,
				InvZ: false,
			},
			acceleration: &Acceleration{
				Y: 0.1,
			},
			expectedAccelerationY: -0.1,
		},
		{
			name: "inverted y and cam y <> y axis mapping",
			axisMap: &AxisMap{
				CamX: "Y",
				CamY: "Z",
				CamZ: "X",
				InvX: false,
				InvY: true,
				InvZ: false,
			},
			acceleration: &Acceleration{
				Z: 0.1,
			},
			expectedAccelerationY: -0.1,
		},
		{
			name: "not inverted and cam z = z axis mapping",
			axisMap: &AxisMap{
				CamX: "X",
				CamY: "Y",
				CamZ: "Z",
				InvX: false,
				InvY: false,
				InvZ: false,
			},
			acceleration: &Acceleration{
				Z: 0.1,
			},
			expectedAccelerationZ: 0.1,
		},
		{
			name: "inverted z and cam z = z axis mapping",
			axisMap: &AxisMap{
				CamX: "X",
				CamY: "Y",
				CamZ: "Z",
				InvX: false,
				InvY: false,
				InvZ: true,
			},
			acceleration: &Acceleration{
				Z: 0.1,
			},
			expectedAccelerationZ: -0.1,
		},
		{
			name: "inverted z and cam z <> z axis mapping",
			axisMap: &AxisMap{
				CamX: "Y",
				CamY: "Z",
				CamZ: "X",
				InvX: false,
				InvY: false,
				InvZ: true,
			},
			acceleration: &Acceleration{
				X: 0.1,
			},
			expectedAccelerationZ: -0.1,
		},
		{
			name: "multiple invert axes and cam axes <> axes mappings",
			axisMap: &AxisMap{
				CamX: "Y",
				CamY: "Z",
				CamZ: "X",
				InvX: true,
				InvY: true,
				InvZ: false,
			},
			acceleration: &Acceleration{
				X: 1.0,
				Y: 0.1,
				Z: 0.1,
			},
			expectedAccelerationX: -0.1,
			expectedAccelerationY: -0.1,
			expectedAccelerationZ: 1.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedAccelerationX, test.axisMap.X(test.acceleration))
			assert.Equal(t, test.expectedAccelerationY, test.axisMap.Y(test.acceleration))
			assert.Equal(t, test.expectedAccelerationY, test.axisMap.Y(test.acceleration))
		})
	}
}
