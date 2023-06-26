package main

import (
	"github.com/streamingfast/imu-controller/device/iim42652"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ParseAxisMap(t *testing.T) {
	tests := []struct {
		name            string
		axisMap         string
		expectedAxisMap *iim42652.AxisMap
		expectedError   bool
	}{
		{
			name:          "missing comma",
			axisMap:       "CamX:X CamY:Y CamZ:Z",
			expectedError: true,
		},
		{
			name:          "missing colon",
			axisMap:       "CamX-X,CamY-Y,CamZ-Z",
			expectedError: true,
		},
		{
			name:          "axis map contains more than 1 colon",
			axisMap:       "CamX:X:X,CamY:Y:Y,CamZ:Z:Z",
			expectedError: true,
		},
		{
			name:    "valid axis mapping",
			axisMap: "CamX:Y,CamY:Z,CamZ:X",
			expectedAxisMap: &iim42652.AxisMap{
				CamX: "Y",
				CamY: "Z",
				CamZ: "X",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			axisMap, err := parseAxisMap(test.axisMap)
			if err != nil {
				if test.expectedError {
					assert.True(t, true)
					return
				}
				assert.True(t, false)
				return
			}
			assert.Equal(t, test.expectedAxisMap, axisMap)
		})
	}
}

func Test_ParseInvertedAxes(t *testing.T) {
	tests := []struct {
		name               string
		invertedAxisMap    string
		initialAxisMapping *iim42652.AxisMap
		expectedAxisMap    *iim42652.AxisMap
		expectedError      bool
	}{
		{
			name:            "missing comma",
			invertedAxisMap: "X:false Y:false Z:false",
			expectedError:   true,
		},
		{
			name:            "missing colon",
			invertedAxisMap: "X-false,Y-false,Z-false",
			expectedError:   true,
		},
		{
			name:            "inverted axis map contains more than 1 colon",
			invertedAxisMap: "X:false:false,Y:false:false,Z:false:false",
			expectedError:   true,
		},
		{
			name:            "valid inverted axis mapping",
			invertedAxisMap: "X:false,Y:true,Z:false",
			initialAxisMapping: &iim42652.AxisMap{
				CamX: "X",
				CamY: "Y",
				CamZ: "Z",
			},
			expectedAxisMap: &iim42652.AxisMap{
				CamX: "X",
				CamY: "Y",
				CamZ: "Z",
				InvX: false,
				InvY: true,
				InvZ: false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invX, invY, invZ, err := parseInvertedMappings(test.invertedAxisMap)
			if err != nil {
				if test.expectedError {
					assert.True(t, true)
					return
				}
				assert.True(t, false)
				return
			}
			test.initialAxisMapping.SetInvertedAxes(invX, invY, invZ)
			assert.Equal(t, test.expectedAxisMap, test.initialAxisMapping)
		})
	}
}
