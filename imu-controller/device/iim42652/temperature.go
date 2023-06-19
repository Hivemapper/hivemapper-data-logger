package iim42652

import "fmt"

type Temperature *float64

func NewTemperature(t float64) Temperature {
	temp := new(float64)
	*temp = t
	return temp
}

func (i *IIM42652) GetTemperature() (Temperature, error) {
	i.registerLock.Lock()
	defer i.registerLock.Unlock()

	err := i.setBank(RegisterTemperatureData.Bank)
	if err != nil {
		return nil, fmt.Errorf("setting bank %s: %w", RegisterTemperatureData.Bank.String(), err)
	}

	msg := make([]byte, 7)
	result := make([]byte, 7)
	msg[0] = ReadMask | byte(RegisterTemperatureData.Address)
	if err := i.connection.Tx(msg, result); err != nil {
		return nil, fmt.Errorf("reading to SPI port: %w", err)
	}

	temp := int16(result[1])<<8 | int16(result[2])
	var val *float64
	val = new(float64)
	*val = float64(temp)/132.48 + 25

	return val, nil
}
