package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Hivemapper/hivemapper-data-logger/imu-controller/device/iim42652"
)

var (
	cameraType = flag.String("camera-type", "hdcs", "Camera type ('hdc' or 'hdcs' only options for now)")
	outputFile = flag.String("output-file", "output.txt", "File to record the data")
)

type IMUData struct {
	Timestamp    time.Time  `json:"timestamp"`
	Acceleration [3]float64 `json:"acceleration"`
	AngularRate  [3]float64 `json:"angular_rate"`
	Temperature  float64    `json:"temperature"`
}

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println("Usage: program <device path>")
		return
	}

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

	file, err := os.Create(*outputFile)
	if err != nil {
		panic(fmt.Errorf("creating output file: %w", err))
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

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

		data := IMUData{
			Timestamp:    time.Now(),
			Acceleration: [3]float64{acceleration.X, acceleration.Y, acceleration.Z},
			AngularRate:  [3]float64{angularRate.X, angularRate.Y, angularRate.Z},
			Temperature:  *temperature, // Dereference the pointer to get the float64 value
		}

		if err := encoder.Encode(&data); err != nil {
			panic(fmt.Errorf("writing to file: %w", err))
		}

		fmt.Printf("%+v\n", data)
	}
}
