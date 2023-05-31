package imu

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	ContinuousCountWindow        int     `json:"continuous_count_window"`
	StopEndContinuousCountWindow int     `json:"stop_end_continuous_count_window"`
	MinimumMagnitudeThreshold    float64 `json:"minimum_magnitude_threshold"`
	LeftTurnThreshold            float64 `json:"left_turn_threshold"`
	RightTurnThreshold           float64 `json:"right_turn_threshold"`
	GForceAcceleratorThreshold   float64 `json:"g_force_accelerator_threshold"`
	GForceDeceleratorThreshold   float64 `json:"g_force_decelerator_threshold"`
}

func (c *Config) String() string {
	j, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(j)
}

func LoadConfig(filename string) *Config {
	var conf *Config
	jsonFile, err := os.Open(filename)
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("can't read imu-logger.json file, using default config\n")
		conf = DefaultConfig()
	}

	if len(byteValue) > 0 {
		err = json.Unmarshal(byteValue, &conf)
		if err != nil {
			fmt.Printf("imu-logger json config file is invalid, using default config\n")
			conf = DefaultConfig()
		}
	}
	return conf
}

func DefaultConfig() *Config {
	return &Config{
		ContinuousCountWindow:        10,
		StopEndContinuousCountWindow: 100,
		MinimumMagnitudeThreshold:    0.2,
		LeftTurnThreshold:            0.15,
		RightTurnThreshold:           -0.15,
		GForceAcceleratorThreshold:   0.25,
		GForceDeceleratorThreshold:   -0.25,
	}
}
