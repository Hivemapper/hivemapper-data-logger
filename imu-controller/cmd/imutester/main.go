package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/streamingfast/imu-controller/device/iim42652"
)

var skipPwrMngt = flag.Bool("skip-power-management", false, "skip power management")

func main() {
	flag.Parse()
	devPath := os.Args[1]

	imuDevice := iim42652.NewSpi(
		devPath,
		iim42652.AccelerationSensitivityG16,
		iim42652.GyroScalesG2000,
		true,
		*skipPwrMngt,
	)

	err := imuDevice.Init()
	if err != nil {
		panic(fmt.Errorf("initializing IMU: %w", err))
	}

	for {
		time.Sleep(10 * time.Millisecond)
		acceleration, err := imuDevice.GetAcceleration()
		if err != nil {
			panic(fmt.Errorf("getting acceleration: %w", err))
		}

		angularRate, err := imuDevice.GetGyroscopeData()
		if err != nil {
			panic(fmt.Errorf("getting angular rate: %w", err))
		}

		temperature, err := imuDevice.GetTemperature()
		if err != nil {
			panic(fmt.Errorf("getting temperature: %w", err))
		}

		fmt.Println("< -- >")
		fmt.Println("acceleration:", acceleration)
		fmt.Println("angularRate:", angularRate)
		fmt.Println("temperature:", temperature)

	}
}
