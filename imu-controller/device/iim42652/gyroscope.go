package iim42652

import (
	"fmt"
)

type GyroScale float64

const (
	GyroScalesG2000  GyroScale = 2000.0 / ShortMax
	GyroScalesG1000  GyroScale = 1000.0 / ShortMax
	GyroScalesG500   GyroScale = 500.0 / ShortMax
	GyroScalesG250   GyroScale = 250.0 / ShortMax
	GyroScalesG125   GyroScale = 125.0 / ShortMax
	GyroScalesG62_5  GyroScale = 62.5 / ShortMax
	GyroScalesG31_25 GyroScale = 31.25 / ShortMax
	GyroScalesG15_62 GyroScale = 15.62 / ShortMax
)

type AngularRate struct {
	RawX int16
	RawY int16
	RawZ int16
	X    float64
	Y    float64
	Z    float64
}

func NewGyroscope(x, y, z int16, gyroScale GyroScale) *AngularRate {
	return &AngularRate{
		RawX: x,
		RawY: y,
		RawZ: z,
		X:    float64(x) * float64(gyroScale),
		Y:    float64(y) * float64(gyroScale),
		Z:    float64(z) * float64(gyroScale),
	}
}

func (a *AngularRate) String() string {
	return fmt.Sprintf("AngularRate: camX:%.5f, camY:%.5f, camZ: %.5f", a.CamX(), a.CamY(), a.CamZ())
}

func (a *AngularRate) CamX() float64 {
	return a.Z
}

func (a *AngularRate) CamY() float64 {
	return a.X
}

func (a *AngularRate) CamZ() float64 {
	return a.Y
}

/*
- AngularRate Start-Up Time -> Time from gyro enable to gyro drive ready: 30 (seconds? ms?)
- 4.6 Three-Axis Mems AngularRate with 16-Bit ADCS and signal ConditioningWhen the gyroscope is rotated about any of the sense axes, the Coriolis Effect causes a
  vibration that is detected by a capacitive pickoff

- GYRO_DATA_X1 -> Upper byte of Gyro X-axis data (15:8) Addr: 37
- GYRO_DATA_X0 -> Lower byte of Gyro X-axis data (7:0) Addr: 38
- GYRO_DATA_Y1 -> Upper byte of Gyro Y-axis data (15:8) Addr: 39
- GYRO_DATA_Y0 -> Lower byte of Gyro Y-axis data (7:0) Addr: 40
- GYRO_DATA_Z1 -> Upper byte of Gyro Z-axis data (15:8) Addr: 41
- GYRO_DATA_Z0 -> Lower byte of Gyro Z-axis data (7:0) Addr: 42
*/

func (i *IIM42652) GetGyroscopeData() (*AngularRate, error) {
	i.registerLock.Lock()
	defer i.registerLock.Unlock()

	err := i.setBank(RegisterGyroscopeDataX1.Bank)
	if err != nil {
		return nil, fmt.Errorf("setting bank %s: %w", RegisterGyroscopeDataX1.Bank.String(), err)
	}

	msg := make([]byte, 7)
	result := make([]byte, 7)
	msg[0] = ReadMask | byte(RegisterGyroscopeDataX1.Address)
	if err := i.connection.Tx(msg, result); err != nil {
		return nil, fmt.Errorf("reading to SPI port: %w", err)
	}

	x := int16(result[1])<<8 | int16(result[2])
	y := int16(result[3])<<8 | int16(result[4])
	z := int16(result[5])<<8 | int16(result[6])

	return NewGyroscope(x, y, z, i.gyroScale), nil
}
