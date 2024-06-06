package iim42652

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Hivemapper/hivemapper-data-logger/logger"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
	_ "periph.io/x/host/v3"
)

type IIM42652 struct {
	deviceName              string
	port                    spi.PortCloser
	connection              spi.Conn
	currentBank             Bank
	accelerationSensitivity AccelerationSensitivity
	gyroScale               GyroScale
	DeviceType              logger.CamType

	debug bool
}

func NewSpi(device string, accelerationSensitivity AccelerationSensitivity, gyroScale GyroScale, debug bool, deviceType string) *IIM42652 {
	cam, err := logger.GetEnumFromString(deviceType)
	if err != nil {
		fmt.Println("Error getting device type: ", err)
	}
	return &IIM42652{
		deviceName:              device,
		accelerationSensitivity: accelerationSensitivity,
		gyroScale:               gyroScale,
		DeviceType:              cam,
		debug:                   debug,
	}
}

func (i *IIM42652) Init() error {
	if state, err := host.Init(); err != nil {
		return fmt.Errorf("failed to initialize driver: %w", err)
	} else {
		fmt.Println("driver state:", state)
	}

	refs := spireg.All()
	fmt.Println("SPI ports available:", len(refs))
	for _, ref := range refs {
		fmt.Println("SPI:", ref.Name, ref.Number, ref.Aliases)
	}
	// Use spireg SPI port registry to find the first available SPI bus.
	p, err := spireg.Open(i.deviceName)
	if err != nil {
		return fmt.Errorf("openning SPI port with name %q: %w", i.deviceName, err)
	}
	i.port = p

	// Convert the spi.Port into a spi.Conn so it can be used for communication.
	c, err := p.Connect(24000000*physic.Hertz, spi.Mode0, 8)
	if err != nil {
		return fmt.Errorf("connecting to SPI port %q: %w", i.deviceName, err)
	}
	i.connection = c
	fmt.Println("SPI port connected:", i.deviceName)

	// TODO: Implement GPIO interrupt if FIFO use is desired
	// // Configure the GPIO pin for interrupts
	// fmt.Println(spireg.All())
	// fmt.Println(gpioreg.All())
	// pinName := fmt.Sprintf("GPIO%d", GPIO_INT1)
	// i.interruptPin = gpioreg.ByName(pinName)
	// if i.interruptPin == nil {
	// 	return fmt.Errorf("failed to find GPIO pin %d", GPIO_INT1)
	// }

	// if err := i.interruptPin.In(gpio.PullUp, gpio.FallingEdge); err != nil {
	// 	return fmt.Errorf("failed to set pin to input mode: %w", err)
	// }

	// Set up rest of IMU
	err = i.SetupPower(GyroModeLowNoise | AccelerometerModeLowNoise)
	if err != nil {
		return fmt.Errorf("setting up power: %w", err)
	}
	// Keep as default SPI mode 0 (0x00)
	err = i.WriteRegister(RegisterDeviceConfig, SPI_MODE_O)
	if err != nil {
		return fmt.Errorf("setting deviceConfig: %w", err)
	}
	time.Sleep(time.Second)

	deviceConfig, err := i.ReadRegister(RegisterDeviceConfig)
	if err != nil {
		return fmt.Errorf("getting deviceConfig: %w", err)
	}
	fmt.Println("IMU: deviceConfig->", hex.EncodeToString([]byte{deviceConfig}))

	// Write SPI_SLEW_RATE
	// err = i.WriteRegister(RegisterDriveConfig, SPI_SLEW_RATE)
	// if err != nil {
	// 	return fmt.Errorf("setting driveConfig: %w", err)
	// }
	// time.Sleep(time.Second)

	driveConfig, err := i.ReadRegister(RegisterDriveConfig)
	if err != nil {
		return fmt.Errorf("getting driveConfig: %w", err)
	}
	fmt.Println("IMU: driveConfig->", hex.EncodeToString([]byte{driveConfig}))

	// Calibrate and set gyro
	i.CalibrateGyro(GYRO_CALIBRATION_MAX_SAMPLES)

	// Set ODR of gyro
	gyroODR, err := i.ReadRegister(RegisterGyroscopeConfig0)
	if err != nil {
		return fmt.Errorf("getting gyroODR: %w", err)
	}
	if gyroODR != GYRO_ODR {
		err = i.WriteRegister(RegisterGyroscopeConfig0, GYRO_ODR)
		if err != nil {
			return fmt.Errorf("setting gyroODR: %w", err)
		}
		fmt.Println("gyroODR:", hex.EncodeToString([]byte{gyroODR}))
	}

	// Set ODR of accel
	accelODR, err := i.ReadRegister(RegisterAccelConfig)
	if err != nil {
		return fmt.Errorf("getting accelODR: %w", err)
	}
	if accelODR != ACCEL_ODR {
		err = i.WriteRegister(RegisterAccelConfig, ACCEL_ODR)
		if err != nil {
			return fmt.Errorf("setting accelODR: %w", err)
		}
	}

	// Check ODR values here
	gyroODR, err = i.ReadRegister(RegisterGyroscopeConfig0)
	if err != nil {
		return fmt.Errorf("getting gyroODR: %w", err)
	}
	fmt.Println("gyroODR:", hex.EncodeToString([]byte{gyroODR}))

	accelODR, err = i.ReadRegister(RegisterAccelConfig)
	if err != nil {
		return fmt.Errorf("getting accelODR: %w", err)
	}
	fmt.Println("accelODR:", hex.EncodeToString([]byte{accelODR}))

	return nil
}

func (i *IIM42652) SetupPower(pwrMode byte) error {
	err := i.WriteRegister(RegisterPwrMgmt0, pwrMode)
	if err != nil {
		return fmt.Errorf("enabling Gyro and Acc: %w", err)
	}
	time.Sleep(time.Millisecond)

	pwrManagement, err := i.ReadRegister(RegisterPwrMgmt0)
	if err != nil {
		return fmt.Errorf("getting pwrManagement: %w", err)
	}
	if pwrManagement == GyroModeLowNoise|AccelerometerModeLowNoise {
		fmt.Println("IMU devices powered on in Low Noise Mode! ")
	} else {
		return fmt.Errorf("failed to power on IMU devices")
	}
	return nil
}

func (i *IIM42652) ResetSignalPath() error {
	err := i.WriteRegister(RegisterSignalPathReset, 0xFF)
	if err != nil {
		return fmt.Errorf("resetting signal path: %w", err)
	}
	time.Sleep(time.Second)
	return nil
}

func (i *IIM42652) Close() error {
	return i.port.Close()
}

func (i *IIM42652) setBank(b Bank) error {
	if b == i.currentBank {
		return nil
	}

	err := i.connection.Tx([]byte{byte(RegisterBankSel.Address), byte(b)}, nil)
	if err != nil {
		return fmt.Errorf("setting bank: %w", err)
	}
	i.currentBank = b
	time.Sleep(time.Millisecond)
	return nil
}

func (i *IIM42652) WriteRegister(reg *Register, value byte) error {

	i.Debugf("Writing bank %q, reg %q: %s\n", reg.Bank, reg.Address, hex.EncodeToString([]byte{value}))

	err := i.setBank(reg.Bank)
	if err != nil {
		return fmt.Errorf("setting bank: %w", err)
	}

	msg := []byte{byte(reg.Address), value}
	if err := i.connection.Tx(msg, nil); err != nil {
		return fmt.Errorf("writing reg %q: %w", hex.EncodeToString(msg), err)
	}
	return nil
}

func (i *IIM42652) ReadRegister(reg *Register) (result byte, err error) {

	err = i.setBank(reg.Bank)
	if err != nil {
		return 0, fmt.Errorf("setting bank: %w", err)
	}

	msg := make([]byte, 2)
	msg[0] = ReadMask | byte(reg.Address)
	r := make([]byte, 2)

	if err := i.connection.Tx(msg, r); err != nil {
		return 0x0, fmt.Errorf("writing to SPI port: %w", err)
	}
	result = r[1]
	return result, nil
}

func (i *IIM42652) UpdateRegister(reg *Register, update func(currentValue byte) byte) error {
	err := i.setBank(reg.Bank)
	if err != nil {
		return fmt.Errorf("setting Bank0: %w", err)
	}

	d, err := i.ReadRegister(reg)
	if err != nil {
		return fmt.Errorf("reading from reg %q: %w", reg, err)
	}
	fmt.Printf("Read %q: %s\n", reg, hex.EncodeToString([]byte{d}))
	d = update(d)
	fmt.Printf("Writing! %q: %s\n", reg, hex.EncodeToString([]byte{d}))
	err = i.WriteRegister(reg, d)
	if err != nil {
		return fmt.Errorf("writing to reg %q: %w", reg, err)
	}
	return nil
}

func (i *IIM42652) ReadInt16(regs ...*Register) (int16, error) {
	if len(regs) != 2 {
		return 0, fmt.Errorf("only 2 bytes per read")
	}

	h, err := i.ReadRegister(regs[0])
	if err != nil {
		return 0, fmt.Errorf("reading high byte: %w", err)
	}

	l, err := i.ReadRegister(regs[1])
	if err != nil {
		return 0, fmt.Errorf("reading low byte: %w", err)
	}

	return int16(h)<<8 | int16(l), nil
}

func (i *IIM42652) Debugln(a ...any) {
	if i.debug {
		fmt.Println(a...)
	}
}

func (i *IIM42652) Debugf(format string, a ...any) {
	if i.debug {
		fmt.Printf(format, a...)
	}
}
