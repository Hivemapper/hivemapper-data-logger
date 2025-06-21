package iim42652

import (
	"fmt"
)

type FifoImuRawData struct {
	Acceleration *Acceleration
	AngularRate  *AngularRate
	Temperature  Temperature
	Fsync        *Fsync
}

func (i *IIM42652) GetFifo() ([]FifoImuRawData, error) {
	i.registerLock.Lock()
	defer i.registerLock.Unlock()

	err := i.setBank(RegisterFifoCountH.Bank)
	if err != nil {
		return nil, fmt.Errorf("setting bank %s: %w", RegisterFifoCountH.Bank.String(), err)
	}

	msg := make([]byte, 7)
	result := make([]byte, 7)
	msg[0] = ReadMask | byte(RegisterFifoCountH.Address)
	if err := i.connection.Tx(msg, result); err != nil {
		return nil, fmt.Errorf("reading to SPI port: %w", err)
	}

	fifoCount := int16(result[1])<<8 | int16(result[2])
	// fmt.Println("Fifo count:", fifoCount)

	// ready fifo data for the number of samples in fifoCount
	err = i.setBank(RegisterFifoData.Bank)
	if err != nil {
		return nil, fmt.Errorf("setting bank %s: %w", RegisterFifoData.Bank.String(), err)
	}

	dataList := make([]FifoImuRawData, 0, fifoCount)

	for jj := int16(0); jj < fifoCount; jj++ {
		msg = make([]byte, 17)
		result = make([]byte, 17)
		msg[0] = ReadMask | byte(RegisterFifoData.Address)
		if err = i.connection.Tx(msg, result); err != nil {
			return nil, fmt.Errorf("reading fifo data to SPI port: %w", err)
		}

		x := int16(result[2])<<8 | int16(result[3])
		y := int16(result[4])<<8 | int16(result[5])
		z := int16(result[6])<<8 | int16(result[7])
		acceleration := NewAcceleration(x, y, z, i.accelerationSensitivity)

		gyro_x := int16(result[8])<<8 | int16(result[9])
		gyro_y := int16(result[10])<<8 | int16(result[11])
		gyro_z := int16(result[12])<<8 | int16(result[13])
		gyroscope := NewGyroscope(gyro_x, gyro_y, gyro_z, i.gyroScale)

		temperature := NewTemperature(float64(int16(result[14]))/2.07 + 25)

		time_delta := int16(result[15])<<8 | int16(result[16])
		// read only bit 7 which is the UI_FSYNC_INT

		// read the 2:3 bits (0-indexed) of first byte
		var mask byte = 0b00001100
		fsyncBits := (result[1] & mask) >> 2
		var fsync_int bool = fsyncBits == 0b11

		fsync := NewFsync(time_delta, fsync_int)

		data := FifoImuRawData{
			Acceleration: acceleration,
			AngularRate:  gyroscope,
			Temperature:  temperature,
			Fsync:        fsync,
		}

		dataList = append(dataList, data)

	}

	return dataList, nil
}
