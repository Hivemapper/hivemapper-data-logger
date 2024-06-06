package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Hivemapper/hivemapper-data-logger/imu-controller/device/iim42652"
)

var cameraType = flag.String("camera-type", "hdcs", "Camera type ('hdc' or 'hdcs' only options for now)")

func main() {
	flag.Parse()
	devPath := os.Args[1]

	imuDevice := iim42652.NewSpi(
		devPath,
		iim42652.AccelerationSensitivityG16,
		iim42652.GyroScalesG2000,
		true,
		*cameraType,
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
