package main

import (
	"fmt"

	"github.com/streamingfast/hm-imu-logger/data"
	"github.com/streamingfast/hm-imu-logger/device/iim42652"
	"github.com/streamingfast/hm-imu-logger/tui"
)

func main() {
	imu := iim42652.NewSpi("/dev/spidev0.0", iim42652.AccelerationSensitivityG16, true)
	err := imu.Init()
	if err != nil {
		panic(fmt.Errorf("initializing IMU: %w", err))
	}

	p := data.NewPipeline(imu)

	go func() {
		err := p.Run()
		if err != nil {
			panic(fmt.Errorf("running pipeline: %w", err))
		}
	}()

	app := tui.NewApp(p)
	err = app.Run()
	if err != nil {
		panic(fmt.Errorf("running app: %w", err))
	}

	//for {
	//	intStatus2, err := imu.ReadRegister(iim42652.RegisterIntStatus2)
	//	if err != nil {
	//		panic("failed to read intstatus2")
	//	}
	//
	//	acc, err := imu.GetAcceleration()
	//	if err != nil {
	//		panic("failed to read acceleration")
	//	}
	//
	//	if err != nil {
	//		panic("failed to read whoami")
	//	}
	//	if intStatus2 > 0 {
	//		fmt.Println("Grrrrrrrrr:", intStatus2)
	//		if intStatus2&0x04 == 0x04 {
	//			fmt.Println("Cam X!", intStatus2&0x04 == 0x04)
	//		}
	//		if intStatus2&0x01 == 0x01 {
	//			fmt.Println("Cam Y!", intStatus2&0x01 == 0x01)
	//		}
	//		if intStatus2&0x02 == 0x02 {
	//			fmt.Println("Cam Z!", intStatus2&0x02 == 0x02)
	//		}
	//		//fmt.Println("Cam X:", intStatus2&0x04 == 0x04, "Y:", intStatus2&0x01 == 0x01, "Z:", intStatus2&0x02 == 0x02, "SMD:", intStatus2&0x08 == 0x08)
	//		fmt.Println("Acceleration:", acc)
	//	}
	//	//if intStatus2&0x08 == 0x08 {
	//	//	fmt.Println("Significant motion detected!")
	//	//}
	//	time.Sleep(10 * time.Millisecond)
	//}

	//// TODO: this for loop here would need to send messages to the
	//acceleration, err := imu.GetAcceleration()
	//if err != nil {
	//	panic(fmt.Errorf("getting acceleration: %w", err))
	//}

}
