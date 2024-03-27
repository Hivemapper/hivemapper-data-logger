package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"golang.org/x/exp/io/i2c"
)

// Redundant file, delete later
const I2C_DEVICE = "/dev/i2c-0"
const I2C_ADDRESS = 0x30
const MAX_TRIES = 150
const NULL_FIELD = 524288

func readData(dev *i2c.Device) ([3]float64, error) {
	dev.WriteReg(0x1B, []byte{0x21})
	i := 0
	for i = 0; i < MAX_TRIES; i++ {
		result := [1]byte{}
		dev.ReadReg(0x18, result[:])
		fmt.Printf("result %v\n", result)

		if result[0]&(1<<6) != 0 {
			break
		}
	}
	if i == MAX_TRIES {
		return [3]float64{}, fmt.Errorf("failed to measure data")
	}

	data := [9]byte{}
	dev.ReadReg(0x00, data[:])
	fmt.Printf("data %v\n", data)

	mag_x := float64(int(binary.BigEndian.Uint32([]byte{0, data[0], data[1], data[6]}))>>4-NULL_FIELD) / 16384 * 1000
	mag_y := float64(int(binary.BigEndian.Uint32([]byte{0, data[2], data[3], data[7]}))>>4-NULL_FIELD) / 16384 * 1000
	mag_z := float64(int(binary.BigEndian.Uint32([]byte{0, data[4], data[5], data[8]}))>>4-NULL_FIELD) / 16384 * 1000
	return [3]float64{mag_x, mag_y, mag_z}, nil
}

func readI2C() {
	fmt.Println("readi2c")

	// Open a connection to the I2C device.
	dev, err := i2c.Open(&i2c.Devfs{Dev: I2C_DEVICE}, I2C_ADDRESS) // Change the address according to your device.
	if err != nil {
		log.Fatalf("Failed to open I2C device: %v", err)
	}
	defer dev.Close()

	for {
		res, err := readData(dev)
		if err != nil {
			log.Fatalf("readData: %v", err)
		}

		// Process the received data.
		// Here you can do whatever processing is necessary with the data read from the device.
		log.Printf("Read data: %v\n", res)

		// Delay or perform other operations as needed.
		time.Sleep(100 * time.Millisecond)
	}
}
