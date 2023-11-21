/*
imucalibrator calibrates IIM42652's sensors.

Usage:

	imucalibrator [flags]

The flags are:

	--sensor
		Sensor to calibrate. This is a required flag. Possible values are
		'gyro' and 'accelerometer'
	--dev-path
		Path to the spi device. By default, this is '/dev/spidev0.0'
	--max-samples int
		The maximum number of samples to take for calibration. Default is 200
	--clear-calibration
		Clears existing calibration data from the sensor set by the sensor flag.
	--verify-calibration
		Verify that measured values make sense.

imucalibrator takes a sample of imu sensor data, averages it, then programs the imu
user register to use the calculated average as a bias.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/streamingfast/imu-controller/device/iim42652"
)

var (
	sensor            = flag.String("sensor", "", "The sensor to calibrate or clear. Required. Values: gyro,accelerometer")
	devicePath        = flag.String("dev-path", "/dev/spidev0.0", "The dev path of the spi device. Default is /dev/spidev0.0")
	maxSamples        = flag.Int("max-samples", 200, "The maximum number of samples to take for calibration. Default is 200")
	clearCalibration  = flag.Bool("clear-calibration", false, "Clear existing calibration data from the IMU")
	verifyCalibration = flag.Bool("verify-calibration", false, "Verify that measured values make sense.")
)

func abs(value int32) int32 {
	if value >= 0 {
		return value
	}
	return -value
}

// Found by experimenting.
const (
	accelerometerVerificationCutoff int32 = 15
	gyroVerificationCutoff          int32 = 7
)

func calibrateGyro(imuDevice *iim42652.IIM42652) error {
	bias, err := imuDevice.CalibrateGyro(int32(*maxSamples))
	if err != nil {
		return fmt.Errorf("calibrating gyro: %w", err)
	}
	fmt.Println("gyro calibration values:", bias)

	time.Sleep(60 * time.Millisecond)
	result, err := verifyGyro(imuDevice)
	if err != nil {
		return err
	}
	if !result {
		fmt.Println("Gyro values were not in expected range. Consider clearing and recalibrating.")
	}
	return nil
}

func verifyGyro(imuDevice *iim42652.IIM42652) (result bool, err error) {
	average, err := imuDevice.AverageGyroSensorOutput(int32(*maxSamples))
	if err != nil {
		return false, err
	}

	fmt.Printf("verifyGyro: raw reading {X: %v Y: %v Z: %v}\n", average[0], average[1], average[2])
	cutoff := gyroVerificationCutoff
	if abs(average[0]) >= cutoff || abs(average[1]) >= cutoff || abs(average[2]) >= cutoff {
		return false, nil
	}
	return true, nil
}

func calibrateAccelerometer(imuDevice *iim42652.IIM42652) error {
	bias, err := imuDevice.CalibrateAccelerometer(int32(*maxSamples))
	if err != nil {
		return fmt.Errorf("calibrating accelerometer: %w", err)
	}
	fmt.Println("accel calibration values:", bias)

	time.Sleep(60 * time.Millisecond)
	result, err := verifyAccelerometer(imuDevice)
	if err != nil {
		return err
	}
	if !result {
		fmt.Println("Accelerometer values were not in expected range. Consider clearing and recalibrating.")
	}
	return nil
}

func verifyAccelerometer(imuDevice *iim42652.IIM42652) (result bool, err error) {
	average, err := imuDevice.AverageAccelerometerSensorOutput(int32(*maxSamples))
	if err != nil {
		return false, err
	}

	fmt.Printf("verifyAccelerometer: raw reading {X: %v Y: %v Z: %v}\n", average[0], average[1], average[2])
	cutoff := accelerometerVerificationCutoff
	if abs(average[0]) >= cutoff || abs(average[1]) >= cutoff || abs(average[2]) >= cutoff {
		return false, nil
	}
	return true, nil
}

func validateFlags() error {
	if *sensor == "" {
		return fmt.Errorf("sensor flag is required")
	}
	if *sensor != "gyro" && *sensor != "accelerometer" {
		return fmt.Errorf("sensor '%v' not recognized, must be 'gyro' or 'accelerometer'", *sensor)
	}
	return nil
}

func main() {
	flag.Parse()

	if err := validateFlags(); err != nil {
		panic(fmt.Errorf("validateflags: %w", err))
	}

	// Note: Only 16G and 2000dps work. iim42652 module does not
	// configure these parameters on the device and relies on the defaults.
	imuDevice := iim42652.NewSpi(
		*devicePath,
		iim42652.AccelerationSensitivityG16,
		iim42652.GyroScalesG2000,
		true,
		false, // skip power management
	)

	err := imuDevice.Init()
	if err != nil {
		panic(fmt.Errorf("initializing IMU: %w", err))
	}

	if *sensor == "gyro" {
		if *clearCalibration {
			err := imuDevice.ClearGyroBias()
			if err != nil {
				panic(fmt.Errorf("clearing IMU: %w", err))
			}
			fmt.Println("Gyro cleared!")
		} else if *verifyCalibration {
			result, err := verifyGyro(imuDevice)
			if err != nil {
				panic(fmt.Errorf("verifying Gyro: %w", err))
			}
			if !result {
				fmt.Println("Gyro verification failed!")
				os.Exit(1)
			}
			fmt.Println("Gyro verified!")
		} else {
			err := calibrateGyro(imuDevice)
			if err != nil {
				panic(fmt.Errorf("calibrating IMU: %w", err))
			}
			fmt.Println("Gyro calibrated!")
		}
	} else {
		if *clearCalibration {
			err := imuDevice.ClearAccelerometerBias()
			if err != nil {
				panic(fmt.Errorf("clearing IMU: %w", err))
			}
			fmt.Println("Accelerometer cleared!")
		} else if *verifyCalibration {
			result, err := verifyAccelerometer(imuDevice)
			if err != nil {
				panic(fmt.Errorf("verifying Accelerometer: %w", err))
			}
			if !result {
				fmt.Println("Accelerometer verification failed!")
				os.Exit(1)
			}
			fmt.Println("Accelerometer verified!")
		} else {
			err := calibrateAccelerometer(imuDevice)
			if err != nil {
				panic(fmt.Errorf("calibrating IMU: %w", err))
			}
			fmt.Println("Accelerometer calibrated!")
		}

	}
}
