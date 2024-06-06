package magnetometer

import (
	"encoding/binary"
	"fmt"
	"time"

	"golang.org/x/exp/io/i2c"
)

const I2C_DEVICE = "/dev/i2c-0"
const I2C_ADDRESS = 0x30
const MAX_TRIES = 150
const NULL_FIELD = 524288

type RawFeed struct {
	device  *i2c.Device
	handler RawFeedHandler
}

func NewRawFeed(handlers RawFeedHandler) *RawFeed {
	return &RawFeed{
		handler: handlers,
	}
}

type RawFeedHandler func(system_time time.Time, mag_x float64, mag_y float64, mag_z float64) error

func readData(dev *i2c.Device) ([3]float64, error) {
	dev.WriteReg(0x1B, []byte{0x21})
	i := 0
	for i = 0; i < MAX_TRIES; i++ {
		result := [1]byte{}
		dev.ReadReg(0x18, result[:])

		if result[0]&(1<<6) != 0 {
			break
		}
	}
	if i == MAX_TRIES {
		return [3]float64{}, fmt.Errorf("failed to measure data")
	}

	data := [9]byte{}
	dev.ReadReg(0x00, data[:])

	mag_x := float64(int(binary.BigEndian.Uint32([]byte{0, data[0], data[1], data[6]}))>>4-NULL_FIELD) / 16384 * 1000
	mag_y := float64(int(binary.BigEndian.Uint32([]byte{0, data[2], data[3], data[7]}))>>4-NULL_FIELD) / 16384 * 1000
	mag_z := float64(int(binary.BigEndian.Uint32([]byte{0, data[4], data[5], data[8]}))>>4-NULL_FIELD) / 16384 * 1000
	return [3]float64{mag_x, mag_y, mag_z}, nil
}

func (f *RawFeed) Init() error {
	// Open a connection to the I2C device.
	dev, err := i2c.Open(&i2c.Devfs{Dev: I2C_DEVICE}, I2C_ADDRESS)
	if err != nil {
		return fmt.Errorf("failed to open I2C device: %v", err)
	}
	f.device = dev
	return nil
}

func (f *RawFeed) Run() error {
	fmt.Println("Run mag feed")
	for {
		time.Sleep(25 * time.Millisecond)
		mag_readings, err := readData(f.device)
		if err != nil {
			return fmt.Errorf("getting magnetometer readings: %w", err)
		}

		err = f.handler(
			time.Now(),
			mag_readings[0],
			mag_readings[1],
			mag_readings[2],
		)
		if err != nil {
			return fmt.Errorf("calling handler: %w", err)
		}

	}
}
