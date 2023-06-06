package main

import (
	"fmt"
	"time"

	"github.com/streamingfast/hivemapper-data-logger/data"

	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"

	"github.com/spf13/cobra"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

var WipCmd = &cobra.Command{
	Use:   "wip",
	Short: "Start the data logger",
	RunE:  wipRun,
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

	// Connect-go
	LogCmd.Flags().String("listen-addr", ":9000", "address to listen on")

	RootCmd.AddCommand(LogCmd)
}

func wipRun(cmd *cobra.Command, args []string) error {
	imuDevice := iim42652.NewSpi("/dev/spidev0.0", iim42652.AccelerationSensitivityG16, iim42652.GyroScalesG2000, true)
	err := imuDevice.Init()
	if err != nil {
		return fmt.Errorf("initializing IMU: %w", err)
	}
	err = imuDevice.UpdateRegister(iim42652.RegisterAccelConfig, func(currentValue byte) byte {
		return currentValue | 0x01
	})
	if err != nil {
		return fmt.Errorf("failed to update register: %w", err)
	}
	conf := imu.LoadConfig(mustGetString(cmd, "imu-config-file"))
	fmt.Println("Config: ", conf.String())

	rawImuEventFeed := imu.NewRawFeed(imuDevice)
	go func() {
		err := rawImuEventFeed.Run()
		if err != nil {
			panic(fmt.Errorf("running raw imu event feed: %w", err))
		}
	}()

	correctedImuEventFeed := imu.NewCorrectedAccelerationFeed()
	go func() {
		err := correctedImuEventFeed.Run(rawImuEventFeed)
		if err != nil {
			panic(fmt.Errorf("running corrected imu event feed: %w", err))
		}
	}()

	serialConfigName := mustGetString(cmd, "gnss-serial-config-name")
	mgaOfflineFilePath := mustGetString(cmd, "gnss-mga-offline-file-path")
	gnssDevice := neom9n.NewNeom9n(serialConfigName, mgaOfflineFilePath)

	gnssEventFeed := gnss.NewEventFeed()
	go func() {
		err := gnssEventFeed.Run(gnssDevice)
		if err != nil {
			panic(fmt.Errorf("running gnss event feed: %w", err))
		}
	}()
	gnssEventSub := gnssEventFeed.Subscribe("merger")
	correctedImuEventSub := correctedImuEventFeed.Subscribe("merger")

	mergedEventFeed := data.NewEventFeedMerger(gnssEventSub, correctedImuEventSub)
	go func() {
		mergedEventFeed.Run()
	}()

	mergedEventSub := mergedEventFeed.Subscribe("wip")

	for {
		select {
		case e := <-mergedEventSub.IncomingEvents:
			fmt.Printf("%T Event: %s", e, e)
		}
	}
}
