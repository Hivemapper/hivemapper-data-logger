package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type Config struct {
	ContinuousCountWindow      int     `json:"continuous_count_window"`
	MinimumMagnitudeThreshold  float64 `json:"minimum_magnitude_threshold"`
	LeftTurnThreshold          float64 `json:"left_turn_threshold"`
	RightTurnThreshold         float64 `json:"right_turn_threshold"`
	GForceAcceleratorThreshold float64 `json:"g_force_accelerator_threshold"`
	GForceDeceleratorThreshold float64 `json:"g_force_decelerator_threshold"`
}

func (c *Config) String() string {
	var strBuilder strings.Builder
	strBuilder.WriteString("\n")
	strBuilder.WriteString(fmt.Sprintf("\tContinuousCountWindow: %d\n", c.ContinuousCountWindow))
	strBuilder.WriteString(fmt.Sprintf("\tMinimumMagnitudeThreshold: %.2f\n", c.MinimumMagnitudeThreshold))
	strBuilder.WriteString(fmt.Sprintf("\tLeftTurnThreshold: %.2f\n", c.LeftTurnThreshold))
	strBuilder.WriteString(fmt.Sprintf("\tRightTurnThreshold: %.2f\n", c.RightTurnThreshold))
	strBuilder.WriteString(fmt.Sprintf("\tGForceAcceleratorThreshold: %.2f\n", c.GForceAcceleratorThreshold))
	strBuilder.WriteString(fmt.Sprintf("\tGForceDeceleratorThreshold: %.2f\n", c.GForceDeceleratorThreshold))
	return strBuilder.String()
}

func LoadConfig(filename string) Config {
	var conf Config
	jsonFile, err := os.Open(filename)
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("can't read imu-logger.json file, using default config\n")
		conf = defaultConfig()
	}

	if len(byteValue) > 0 {
		err = json.Unmarshal(byteValue, &conf)
		if err != nil {
			fmt.Printf("imu-logger json config file is invalid, using default config\n")
			conf = defaultConfig()
		}
	}
	return conf
}

func defaultConfig() Config {
	return Config{
		ContinuousCountWindow:      10,
		MinimumMagnitudeThreshold:  0.2,
		LeftTurnThreshold:          0.15,
		RightTurnThreshold:         -0.15,
		GForceAcceleratorThreshold: 0.25,
		GForceDeceleratorThreshold: -0.25,
	}
}
