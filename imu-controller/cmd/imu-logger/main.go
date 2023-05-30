package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/streamingfast/hm-imu-logger/config"

	"github.com/streamingfast/hm-imu-logger/data"
	"github.com/streamingfast/hm-imu-logger/device/iim42652"
	"github.com/streamingfast/hm-imu-logger/tui"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:          "imu-logger",
	Short:        "Run imu-logger",
	RunE:         imuLoggerRunE,
	Args:         cobra.ExactArgs(0),
	SilenceUsage: true,
}

func init() {
	rootCmd.Flags().String("config-file", "./imu-logger.json", "Path to imu logger config file, defaults to ./imu-logger.json")
}

func imuLoggerRunE(cmd *cobra.Command, args []string) error {
	imu := iim42652.NewSpi("/dev/spidev0.0", iim42652.AccelerationSensitivityG16, true)
	err := imu.Init()
	if err != nil {
		return fmt.Errorf("initializing IMU: %w", err)
	}

	aafDelta, err := imu.ReadRegister(iim42652.RegisterAntiAliasFilterDelta)
	if err != nil {
		return fmt.Errorf("failed to read aafDelta: %w", err)
	}
	fmt.Printf("aafDelt: %b\n", aafDelta)

	affDeltaSqr, err := imu.ReadRegister(iim42652.RegisterAntiAliasFilterDeltaSqr)
	if err != nil {
		return fmt.Errorf("failed to read addDeltaSqr: %w", err)
	}
	fmt.Printf("addDeltaSqr: %b\n", affDeltaSqr)

	affBitshift, err := imu.ReadRegister(iim42652.RegisterAntiAliasFilterBitshift)
	if err != nil {
		return fmt.Errorf("failed to read affBitshift: %w", err)
	}
	fmt.Printf("affBitshift: %b\n", affBitshift)

	p := data.NewPipeline(imu)

	go func() {
		err := p.Run()
		if err != nil {
			panic(fmt.Errorf("running pipeline: %w", err))
		}
	}()

	conf := config.LoadConfig(mustGetString(cmd, "config-file"))
	fmt.Println("Config: ", conf.String())

	eventEmitter := data.NewEventEmitter(conf)
	go func() {
		err := eventEmitter.Run(p)
		if err != nil {
			panic(fmt.Errorf("running event emitter: %w", err))
		}
	}()

	app := tui.NewApp(eventEmitter)
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
