package iim42652

import (
	"fmt"
	"math"
	"time"
)

type AccelerationSensitivity float64

const (
	AccelerationSensitivityG16 AccelerationSensitivity = 16.0 / ShortMax
	AccelerationSensitivityG8  AccelerationSensitivity = 8.0 / ShortMax
	AccelerationSensitivityG4  AccelerationSensitivity = 4.0 / ShortMax
	AccelerationSensitivityG2  AccelerationSensitivity = 2.0 / ShortMax
)

type Acceleration struct {
	RawX           int16
	RawY           int16
	RawZ           int16
	X              float64
	Y              float64
	Z              float64
	TotalMagnitude float64
}

func ProcessNewAcceleration(x, y, z int16, sensitivity AccelerationSensitivity) *Acceleration {
	accel := &Acceleration{
		RawX: x,
		RawY: y,
		RawZ: z,
		X:    float64(x) * float64(sensitivity),
		Y:    float64(y) * float64(sensitivity),
		Z:    float64(z) * float64(sensitivity),
	}

	accel.TotalMagnitude = math.Sqrt(accel.X*accel.X + accel.Y*accel.Y + accel.Z*accel.Z)

	return accel
}

func (a *Acceleration) String() string {
	return fmt.Sprintf("Acceleration{camX:%.5f, camY:%.5f, camZ: %.5f, totalMagn: %.5f}", a.Z, a.X, a.Y, a.TotalMagnitude)
}

func (a *Acceleration) CamX() float64 {
	return a.Z
}

func (a *Acceleration) CamY() float64 {
	return a.X
}

func (a *Acceleration) CamZ() float64 {
	return a.Y
}

//Significant Motion Detection configuration parameters
//	1. WOM_X_TH (Register 0x4Ah in Bank 4)
//	2. WOM_Y_TH (Register 0x4Bh in Bank 4)
//	3. WOM_Z_TH (Register 0x4Ch in Bank 4)
//	4. WOM_INT_MODE (Register 0x57h in Bank 0)
//	5. WOM_MODE (Register 0x57h in Bank 0)
//	6. SMD_MODE (Register 0x57h in Bank 0)
//
//Initialize Sensor in a typical configuration
//----	1. Set accelerometer ODR (Register 0x50h in Bank 0)
//----		ACCEL_ODR = 9 for 50 Hz
//----	2. Set Accel to Low Power mode (Register 0x4Eh in Bank 0) ACCEL_MODE = 2
//----		and (Register 0x4Dh in Bank 0), ACCEL_LP_CLK_SEL = 0, for low power mode 3.
//----	Wait 1 millisecond
//
//
//Initialize APEX
//	1. Set WOM_X_TH to 98 (Register 0x4Ah in Bank 4)
//	2. Set WOM_Y_TH to 98 (Register 0x4Bh in Bank 4)
//	3. Set WOM_Z_TH to 98 (Register 0x4Ch in Bank 4)
//
//	4. Wait 1 millisecond
//
//	5. Enable SMD source for INT1 by setting bit 3 in register INT_SOURCE1 (Register 0x66h in Bank 0) to 1.
//
//		Or if INT2 is selected for SMD, enable SMD source by setting bit 3 in register INT_SOURCE4 (Register
//		0x69h in Bank 0) to 1.
//	6. Wait 50 milliseconds
//	7. Turn on SMD feature by setting WOM_INT_MODE to 0, WOM_MODE to 1, SMD_MODE to 3
//		(Register 0x56h in Bank 0)
//
//
//• Output registers
//		1. Read interrupt register (Register 0x37h in Bank 0) for SMD_INT

func (i *IIM42652) SetupSignificantMotionDetection() error {
	if err := i.WriteRegister(RegisterAccelConfig, 0x09); err != nil {
		return fmt.Errorf("writing to RegisterAccelConfig %q: %w", RegisterAccelConfig, err)
	}

	err := i.UpdateRegister(RegisterPwrMgmt0, func(currentValue byte) byte {
		fmt.Println("currentValue", currentValue, "AccelerometerModeLowPower", AccelerometerModeLowPower, "currentValue | AccelerometerModeLowPower", currentValue|AccelerometerModeLowPower)
		return currentValue | AccelerometerModeLowPower
	})
	if err != nil {
		return fmt.Errorf("updating RegisterPwrMgmt0 %q: %w", RegisterPwrMgmt0, err)
	}

	var AccelerometerLowPowerMode byte = 0x04
	err = i.UpdateRegister(RegisterIntfConfig1, func(currentValue byte) byte {
		if currentValue&AccelerometerLowPowerMode == AccelerometerLowPowerMode {
			currentValue -= AccelerometerLowPowerMode
		}
		return currentValue
	})
	time.Sleep(1 * time.Millisecond)

	//Camera X = IMU Z
	//Canera Y = IMU X
	//Camera Z = IMU Y

	//1g/256=~3.9mg 0.0039

	var fxCamThreshold = 0.15 / 0.0039
	var camXThreshold = byte(fxCamThreshold)

	var fyCamThreshold = 0.15 / 0.0039
	var camYThreshold = byte(fyCamThreshold)

	var fzCamThreshold = 0.250 / 0.0039
	var camZThreshold = byte(fzCamThreshold)

	if err := i.WriteRegister(RegisterAccelWomXThreshold, camYThreshold); err != nil {
		return fmt.Errorf("writing to RegisterAccelWomXThreshold %q: %w", RegisterAccelWomXThreshold, err)
	}

	if err := i.WriteRegister(RegisterAccelWomYThreshold, camZThreshold); err != nil {
		return fmt.Errorf("writing to RegisterAccelWomYThreshold %q: %w", RegisterAccelWomYThreshold, err)
	}

	if err := i.WriteRegister(RegisterAccelWomZThreshold, camXThreshold); err != nil {
		return fmt.Errorf("writing to RegisterAccelWomZThreshold %q: %w", RegisterAccelWomZThreshold, err)
	}
	time.Sleep(1 * time.Millisecond)

	var SmdInt1EN byte = 0x08
	if err := i.UpdateRegister(RegisterIntSource1, func(currentValue byte) byte {
		if currentValue&SmdInt1EN == SmdInt1EN {
			currentValue -= SmdInt1EN
		}
		return currentValue
	}); err != nil {
		return fmt.Errorf("updating RegisterIntSource1 %q: %w", RegisterIntSource1, err)
	}

	time.Sleep(50 * time.Millisecond)

	if err := i.WriteRegister(RegisterSdmConfig0, 0x7); err != nil {
		return fmt.Errorf("writing to RegisterSdmConfig0 %q: %w", RegisterSdmConfig0, err)
	}

	return nil
}

func (i *IIM42652) GetAcceleration() (*Acceleration, error) {
	// i.registerLock.Lock()
	// defer i.registerLock.Unlock()

	err := i.setBank(RegisterAccelDataX1.Bank)
	if err != nil {
		return nil, fmt.Errorf("setting bank %s: %w", RegisterAccelDataX1.Bank.String(), err)
	}

	msg := make([]byte, 7)
	result := make([]byte, 7)
	msg[0] = ReadMask | byte(RegisterAccelDataX1.Address)
	if err := i.connection.Tx(msg, result); err != nil {
		return nil, fmt.Errorf("reading to SPI port: %w", err)
	}

	x := int16(result[1])<<8 | int16(result[2])
	y := int16(result[3])<<8 | int16(result[4])
	z := int16(result[5])<<8 | int16(result[6])

	acc := ProcessNewAcceleration(x, y, z, i.accelerationSensitivity)
	return acc, nil
}
