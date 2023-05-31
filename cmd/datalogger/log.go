package main

import (
	"fmt"
	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/logger"
	"time"

	"github.com/spf13/cobra"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/hivemapper-data-logger/tui"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

var LogCmd = &cobra.Command{
	Use:   "log",
	Short: "Start the data logger",
	RunE:  logRun,
}

func init() {
	// Imu
	LogCmd.Flags().String("imu-config-file", "imu-logger.json", "Imu logger config file. Default path is ./imu-logger.json")

	// GNSS
	LogCmd.Flags().String("gnss-config-file", "gnss-logger.json", "Neom9n logger config file. Default path is ./gnss-logger.json")
	LogCmd.Flags().String("gnss-json-destination-folder", "/mnt/data/gps", "json destination folder")
	LogCmd.Flags().Duration("gnss-json-save-interval", 15*time.Second, "json save interval")
	LogCmd.Flags().Int64("gnss-json-destination-folder-max-size", int64(30000*1024), "json destination folder maximum size") // 30MB
	LogCmd.Flags().String("gnss-serial-config-name", "/dev/ttyAMA1", "Config serial location")
	LogCmd.Flags().String("gnss-mga-offline-file-path", "/mnt/data/mgaoffline.ubx", "path to mga offline files")
	LogCmd.Flags().String("gnss-db-path", "/mnt/data/gnss.v1.0.3.db", "path to sqliteLogger database")
	LogCmd.Flags().Duration("gnss-db-log-ttl", 12*time.Hour, "ttl of logs in database")

	RootCmd.AddCommand(LogCmd)
}

func logRun(cmd *cobra.Command, args []string) error {
	imuDevice := iim42652.NewSpi("/dev/spidev0.0", iim42652.AccelerationSensitivityG16, true)
	err := imuDevice.Init()
	if err != nil {
		return fmt.Errorf("initializing IMU: %w", err)
	}

	serialConfigName := mustGetString(cmd, "gnss-serial-config-name")
	mgaOfflineFilePath := mustGetString(cmd, "gnss-mga-offline-file-path")
	dbPath := mustGetString(cmd, "gnss-db-path")
	logTTl := mustGetDuration(cmd, "gnss-db-log-ttl")
	jsonDestinationFolder := mustGetString(cmd, "gnss-json-destination-folder")
	jsonSaveInterval := mustGetDuration(cmd, "gnss-json-save-interval")
	jsonDestinationFolderMaxSize := mustGetInt64(cmd, "gnss-json-destination-folder-max-size")

	gnssDevice := neom9n.NewNeom9n(serialConfigName, mgaOfflineFilePath)
	sqliteLogger := logger.NewSqlite(dbPath)
	err = sqliteLogger.Init(logTTl)
	if err != nil {
		return fmt.Errorf("initializing sqlite database: %w", err)
	}

	jsonLogger := logger.NewJsonFile(jsonDestinationFolder, jsonDestinationFolderMaxSize, jsonSaveInterval)
	err = jsonLogger.Init()
	if err != nil {
		return fmt.Errorf("initializing json logger database: %w", err)
	}

	// not sure about `RegisterAccelConfig` -> before it was `RegisterAccelConfigStatic2`
	err = imuDevice.UpdateRegister(iim42652.RegisterAccelConfig, func(currentValue byte) byte {
		return currentValue | 0x01
	})

	if err != nil {
		return fmt.Errorf("failed to update register: %w", err)
	}

	// not sure about `RegisterAccelConfig` -> before it was `RegisterAccelConfigStatic2`
	accelAAFstatus, err := imuDevice.ReadRegister(iim42652.RegisterAccelConfig)
	if err != nil {
		return fmt.Errorf("failed to read accelAAFstatus: %w", err)
	}
	fmt.Printf("accelAAFstatus: %b\n", accelAAFstatus)

	aafDelta, err := imuDevice.ReadRegister(iim42652.RegisterAntiAliasFilterDelta)
	if err != nil {
		return fmt.Errorf("failed to read aafDelta: %w", err)
	}
	fmt.Printf("aafDelt: %b\n", aafDelta)

	affDeltaSqr, err := imuDevice.ReadRegister(iim42652.RegisterAntiAliasFilterDeltaSqr)
	if err != nil {
		return fmt.Errorf("failed to read addDeltaSqr: %w", err)
	}
	fmt.Printf("addDeltaSqr: %b\n", affDeltaSqr)

	affBitshift, err := imuDevice.ReadRegister(iim42652.RegisterAntiAliasFilterBitshift)
	if err != nil {
		return fmt.Errorf("failed to read affBitshift: %w", err)
	}
	fmt.Printf("affBitshift: %b\n", affBitshift)

	conf := imu.LoadConfig(mustGetString(cmd, "imu-config-file"))
	fmt.Println("Config: ", conf.String())

	imuEventFeed := imu.NewEventFeed(imuDevice, conf)
	go func() {
		err := imuEventFeed.Run()
		if err != nil {
			panic(fmt.Errorf("running pipeline: %w", err))
		}
	}()

	gnssEventFeed := gnss.NewEventFeed()
	gnssSubscription := gnssEventFeed.Subscribe("tui")

	lastPosition, err := sqliteLogger.GetLastPosition()
	if err != nil {
		return fmt.Errorf("getting last posotion: %w", err)
	}
	err = gnssDevice.Init(lastPosition)
	if err != nil {
		return fmt.Errorf("initializing neom9n: %w", err)
	}

	dataFeed := neom9n.NewDataFeed(gnssEventFeed.HandleData)
	go func() {
		err = gnssDevice.Run(dataFeed, func(now time.Time) {
			dataFeed.SetStartTime(now)
			jsonLogger.StartStoring()
			sqliteLogger.StartStoring()
		})
		if err != nil {
			panic(fmt.Errorf("running gnss: %w", err))
		}
	}()

	//todo: in gnss controller create a device/NEOM9N package and create a struct NEOM9N
	//todo: move code from main to device/NEOM9N
	//todo: create an event feed for gnss device

	//todo: move logger from gnss to this project  under logger package/gnss

	//todo: create a gnss package under data
	//todo: change gnss code to adopt event feed

	//todo: init file logger for imu
	//todo: init file logger for gnss

	//todo: init db logger for imu
	//todo: init db logger for gnss

	//todo: ui is optional and turn off by default

	tuiImuEventSubscription := imuEventFeed.Subscribe("tui")
	app := tui.NewApp(tuiImuEventSubscription, gnssSubscription)
	err = app.Run()
	if err != nil {
		return fmt.Errorf("running app: %w", err)
	}

	return nil
}

func mustGetString(cmd *cobra.Command, flagName string) string {
	val, err := cmd.Flags().GetString(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}

func mustGetDuration(cmd *cobra.Command, flagName string) time.Duration {
	val, err := cmd.Flags().GetDuration(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}

func mustGetInt64(cmd *cobra.Command, flagName string) int64 {
	val, err := cmd.Flags().GetInt64(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}
