package main

import (
	"fmt"

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
	RootCmd.AddCommand(LogCmd)
}

func logRun(cmd *cobra.Command, args []string) error {
	imuDevice := iim42652.NewSpi("/dev/spidev0.0", iim42652.AccelerationSensitivityG16, true)
	err := imuDevice.Init()
	if err != nil {
		return fmt.Errorf("initializing IMU: %w", err)
	}

	//aafDelta, err := imuDevice.ReadRegister(iim42652.RegisterAntiAliasFilterDelta)
	//if err != nil {
	//	return fmt.Errorf("failed to read aafDelta: %w", err)
	//}
	//fmt.Printf("aafDelt: %b\n", aafDelta)
	//
	//affDeltaSqr, err := imuDevice.ReadRegister(iim42652.RegisterAntiAliasFilterDeltaSqr)
	//if err != nil {
	//	return fmt.Errorf("failed to read addDeltaSqr: %w", err)
	//}
	//fmt.Printf("addDeltaSqr: %b\n", affDeltaSqr)
	//
	//affBitshift, err := imuDevice.ReadRegister(iim42652.RegisterAntiAliasFilterBitshift)
	//if err != nil {
	//	return fmt.Errorf("failed to read affBitshift: %w", err)
	//}
	//fmt.Printf("affBitshift: %b\n", affBitshift)

	conf := imu.DefaultConfig()
	//conf := data.config.LoadConfig(mustGetString(cmd, "config-file"))
	//fmt.Println("Config: ", conf.String())

	imuEventFeed := imu.NewEventFeed(imuDevice, conf)
	go func() {
		err := imuEventFeed.Run()
		if err != nil {
			panic(fmt.Errorf("running pipeline: %w", err))
		}
	}()

	tuiImuEventSubscription := imuEventFeed.Subscribe("tui-imu")

	//dbLogger := data.NewLogger(conf, sub)
	//fileLogger := data.NewFileLogger(conf, sub2)

	app := tui.NewApp(tuiImuEventSubscription)
	err = app.Run()
	if err != nil {
		return fmt.Errorf("running app: %w", err)
	}

	return nil
}
