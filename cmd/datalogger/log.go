package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Hivemapper/hivemapper-data-logger/data/gnss"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/Hivemapper/hivemapper-data-logger/data/magnetometer"
	"github.com/Hivemapper/hivemapper-data-logger/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/imu-controller/device/iim42652"
	"github.com/spf13/cobra"
)

var LogCmd = &cobra.Command{
	Use:   "log",
	Short: "Start the data logger",
	RunE:  logRun,
}

func init() {
	// Camera Type
	LogCmd.Flags().String("camera-type", "hdcs", "camera type ('hdc' or 'hdcs' only options for now)")

	// Gnss
	LogCmd.Flags().Int("gnss-initial-baud-rate", 38400, "initial baud rate of gnss device")
	LogCmd.Flags().String("gnss-config-file", "gnss-logger.json", "Neom9n logger config file. Default path is ./gnss-logger.json")
	LogCmd.Flags().String("gnss-dev-path", "/dev/ttyAMA1", "Config serial location")
	LogCmd.Flags().String("gnss-mga-offline-file-path", "/mnt/data/mgaoffline.ubx", "path to mga offline files")
	LogCmd.Flags().Bool("gnss-fix-check", true, "check if gnss fix is set")
	LogCmd.Flags().Bool("gnss-measx-enabled", false, "enable output of MEASX messages")

	// Sqlite database
	LogCmd.Flags().String("db-output-path", "/mnt/data/gnss.v1.1.0.db", "path to sqliteLogger database")
	LogCmd.Flags().Duration("db-log-ttl", 12*time.Hour, "ttl of logs in database")
	LogCmd.Flags().String("imu-dev-path", "/dev/spidev0.0", "Config serial location")

	// Magnetometer
	LogCmd.Flags().Bool("enable-magnetometer", false, "enable reading from magnetometer")

	RootCmd.AddCommand(LogCmd)
}

func logRun(cmd *cobra.Command, _ []string) error {
	fmt.Println(("Entering logRun --------------------------------"))
	// setup section
	serialConfigName := mustGetString(cmd, "gnss-dev-path")
	mgaOfflineFilePath := mustGetString(cmd, "gnss-mga-offline-file-path")
	// Get camera type for axis mapping
	deviceType := mustGetString(cmd, "camera-type")

	dataHandler, err := NewDataHandler(
		mustGetString(cmd, "db-output-path"),
	)
	if err != nil {
		return fmt.Errorf("creating data handler: %w", err)
	}

	// ALL Devices Initialization and Setup
	fmt.Println("IMU:Setup SPI connection to IMU")
	imuDevice := iim42652.NewSpi(
		mustGetString(cmd, "imu-dev-path"),
		iim42652.AccelerationSensitivityG16,
		iim42652.GyroScalesG2000,
		true,
		deviceType,
	)

	fmt.Println("IMU: Running IMU initilization")
	err = imuDevice.Init()
	if err != nil {
		return fmt.Errorf("initializing IMU: %w", err)
	}
	fmt.Println("IMU: IMU initialized")

	gnssDevice := neom9n.NewNeom9n(serialConfigName, mgaOfflineFilePath, mustGetInt(cmd, "gnss-initial-baud-rate"), mustGetBool(cmd, "gnss-measx-enabled"), dataHandler.sqliteLogger.InsertErrorLog)
	err = gnssDevice.Init(nil)
	if err != nil {
		return fmt.Errorf("initializing neom9n: %w", err)
	}

	// Individual Device Feed Creation

	gnssEventFeed := gnss.NewGnssFeed(
		dataHandler.HandlerGnssData,
	)

	rawImuFeed := imu.NewRawFeed(
		imuDevice,
		dataHandler.HandleRawImuFeed,
	)

	// Shared Wait Group for all goroutines in the main loop
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = gnssEventFeed.Run(gnssDevice)
		if err != nil {
			panic(fmt.Errorf("running gnss event feed: %w", err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := rawImuFeed.Run()
		if err != nil {
			panic(fmt.Errorf("running raw imu event feed: %w", err))
		}
	}()

	if mustGetBool(cmd, "enable-magnetometer") {
		magnetometerEventFeed := magnetometer.NewRawFeed(
			dataHandler.HandlerMagnetometerData,
		)

		err = magnetometerEventFeed.Init()
		if err != nil {
			panic(fmt.Errorf("initializing magnetometer feed: %w", err))
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = magnetometerEventFeed.Run()
			if err != nil {
				panic(fmt.Errorf("running magnetometer feed: %w", err))
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	fmt.Println("All goroutines have finished")
	fmt.Println("Main loop has exited")

	return nil
}
