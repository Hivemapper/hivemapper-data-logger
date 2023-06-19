package main

import (
	"fmt"
	"time"

	"github.com/streamingfast/gnss-controller/device/neom9n"
)

func main() {
	serialConfigName := "/dev/ttyAMA1"
	mgaOfflineFilePath := "/mnt/data/mgaoffline.ubx"
	gnssDevice := neom9n.NewNeom9n(serialConfigName, mgaOfflineFilePath)
	err := gnssDevice.Init(nil)
	if err != nil {
		panic(fmt.Errorf("initializing neom9n: %w", err))
	}

	dataFeed := neom9n.NewDataFeed(func(d *neom9n.Data) {
		fmt.Println("Got data", d)
	})

	err = gnssDevice.Run(dataFeed, func(now time.Time) {
		dataFeed.SetStartTime(now)

	})
	if err != nil {
		panic(fmt.Errorf("running gnss device: %w", err))
	}
}
