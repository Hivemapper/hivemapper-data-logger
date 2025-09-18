package imu

import (
	"fmt"
	"os"
	"time"

	"github.com/streamingfast/imu-controller/device/iim42652"
)

type RawFeed struct {
	imu      *iim42652.IIM42652
	handlers []RawFeedHandler
}

type ImuRawData struct {
	acceleration *Acceleration
	angularRate  *iim42652.AngularRate
	temperature  iim42652.Temperature
	fsync        *iim42652.Fsync
}

func NewRawFeed(imu *iim42652.IIM42652, handlers ...RawFeedHandler) *RawFeed {
	return &RawFeed{
		imu:      imu,
		handlers: handlers,
	}
}

type RawFeedHandler func(acceleration *Acceleration, angularRate *iim42652.AngularRate, temperature iim42652.Temperature, fsync *iim42652.Fsync) error

func (f *RawFeed) Run(axisMap *iim42652.AxisMap) error {
	fmt.Println("Run imu raw feed")

	// Open log file once before loop
	logFile, err := os.OpenFile("/data/logger_imu_loop.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer logFile.Close()

	fifoChan := make(chan ImuRawData, 250) //

	go func() {
		for fifodata := range fifoChan {
			for _, handler := range f.handlers {
				if err := handler(fifodata.acceleration, fifodata.angularRate, fifodata.temperature, fifodata.fsync); err != nil {
					fmt.Printf("handler error: %v\n", err)
				}
			}

		}
	}()

	var betweenFsyncs int = -1
	var betweenFsyncsFull int = -1
	var prev_last_packet_time time.Time
	var time_of_last_packet time.Time
	var packet_time time.Time
	var high_freq_counter int = 58
	var low_freq_counter int = 58

	for {

		time_of_last_packet = time.Now().UTC()
		fifopackets, err := f.imu.GetFifo() // Read FIFO data, if needed
		if err != nil {
			return fmt.Errorf("getting fifo data: %w", err)
		}

		if betweenFsyncs == -1 {
			// ignore first set of data
			betweenFsyncs = 1
			betweenFsyncsFull = 1
			prev_last_packet_time = time_of_last_packet

			continue
		}

		validPackets := make([]iim42652.FifoImuRawData, 0, len(fifopackets))
		for _, fifoData := range fifopackets {
			if !fifoData.Fsync.FsyncInt && betweenFsyncs >= 200 {
				betweenFsyncsFull++
				// Discard excess samples (usually only 1)
				continue
			}
			validPackets = append(validPackets, fifoData)

			if fifoData.Fsync.FsyncInt {
				if betweenFsyncs < 200 {
					low_freq_counter++
					if low_freq_counter >= 60 {
						fmt.Println(time.Now().UTC(), "[Warning] ", betweenFsyncs, " samples since last fsync, repeated every 60 instances")
						low_freq_counter = 0
					}
				}
				if betweenFsyncsFull > 200 {
					high_freq_counter++
					if high_freq_counter >= 60 {
						fmt.Println(time.Now().UTC(), "[Warning] ", betweenFsyncsFull, " samples between fsyncs, repeated every 60 instances")
						high_freq_counter = 0
					}
				}
				betweenFsyncs = 1
				betweenFsyncsFull = 1
			} else {
				betweenFsyncs++
				betweenFsyncsFull++
			}
		}

		for packet_idx, fifoData := range validPackets {

			total_packets := len(validPackets)
			if total_packets > 1 && packet_idx < total_packets-1 {
				time_step := time_of_last_packet.Sub(prev_last_packet_time) / time.Duration(total_packets)
				packet_time = prev_last_packet_time.Add(time_step * time.Duration(packet_idx+1))
			} else {
				packet_time = time_of_last_packet
			}

			fifo_raw_data := ImuRawData{
				acceleration: NewAcceleration(axisMap.X(fifoData.Acceleration), axisMap.Y(fifoData.Acceleration), axisMap.Z(fifoData.Acceleration), fifoData.Acceleration.TotalMagnitude, packet_time),
				angularRate:  fifoData.AngularRate,
				temperature:  fifoData.Temperature,
				fsync:        fifoData.Fsync,
			}

			if fifoData.AngularRate.X < -2000.0 {
				fmt.Println("Resetting imu because angular rate is too high:", fifoData.AngularRate.X)
				err := f.imu.Init()
				if err != nil {
					return fmt.Errorf("initializing IMU: %w", err)
				}
			}

			select {
			case fifoChan <- fifo_raw_data:
				// Sent successfully
			default:
				// Channel full, drop or log
				fmt.Println(
					time.Now().UTC(),
					"Warning: fifo data channel full, dropping FIFO data. Current queue size:",
					len(fifoChan),
					"capacity:", cap(fifoChan),
				)
			}
		}

		prev_last_packet_time = time_of_last_packet

		time.Sleep(100 * time.Millisecond)

	}
}
