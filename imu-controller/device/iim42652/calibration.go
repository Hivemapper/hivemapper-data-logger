package iim42652

import (
	"time"
)

////////////////////////////////////////////////////////////
/// CONSTANTS
////////////////////////////////////////////////////////////

// Gyro configuration register constants.
const (
	bitGyroFsSelectPos         byte = 5
	bitGyroConfig0FSSelectMask byte = (0x7 << bitGyroFsSelectPos)
	bitGyroFSSelect2000dps     byte = (0 << bitGyroFsSelectPos)

	bitGyroConfig0ODRpos  byte = 0
	bitGyroConfig0ODRMask byte = (0x0F << bitGyroConfig0ODRpos)
	bitGyroODRSelect1KHz  byte = (0x06 << bitGyroConfig0ODRpos)

	bitGyroConfig1GyroUiFiltOrdPos   byte = 2
	bitGyroConfig1GyroUiFiltOrdMask  byte = (0x3 << bitGyroConfig1GyroUiFiltOrdPos)
	bitGyroConfig1GyroUiFilt3rdOrder byte = (0x2 << bitGyroConfig1GyroUiFiltOrdPos) // 3rd order filter

	bitGyroAccelConfig0GyroFiltPos  byte = 0
	bitGyroAccelConfig0GyroFiltMask byte = (0x0F << bitGyroAccelConfig0GyroFiltPos)
	bitGyroUiFiltBandwidth          byte = (0x04 << bitGyroAccelConfig0GyroFiltPos) // BW_10
)

// User register constants.
const (
	bitGyroXOffuserPosLo  byte = 0
	bitGyroXOffuserMaskLo byte = (0xFF << bitGyroXOffuserPosLo)
	bitGyroXOffuserPosHi  byte = 0
	bitGyroXOffuserMaskHi byte = (0x0F << bitGyroXOffuserPosHi)

	bitGyroYOffuserPosHi  byte = 4
	bitGyroYOffuserMaskHi byte = (0x0F << bitGyroYOffuserPosHi)
	bitGyroYOffuserPosLo  byte = 0
	bitGyroYOffuserMaskLo byte = (0xFF << bitGyroYOffuserPosLo)

	bitGyroZOffuserPosLo  byte = 0
	bitGyroZOffuserMaskLo byte = (0xFF << bitGyroZOffuserPosLo)
	bitGyroZOffuserPosHi  byte = 0
	bitGyroZOffuserMaskHi byte = (0x0F << bitGyroZOffuserPosHi)

	bitAccelXOffuserPosLo  byte = 0
	bitAccelXOffuserMaskLo byte = (0xFF << bitAccelXOffuserPosLo)
	bitAccelXOffuserPosHi  byte = 4
	bitAccelXOffuserMaskHi byte = (0x0F << bitAccelXOffuserPosHi)

	bitAccelYOffuserPosLo  byte = 0
	bitAccelYOffuserMaskLo byte = (0xFF << bitAccelYOffuserPosLo)
	bitAccelYOffuserPosHi  byte = 0
	bitAccelYOffuserMaskHi byte = (0x0F << bitAccelYOffuserPosHi)

	bitAccelZOffuserPosHi  byte = 4
	bitAccelZOffuserMaskHi byte = (0x0F << bitAccelZOffuserPosHi)
	bitAccelZOffuserPosLo  byte = 0
	bitAccelZOffuserMaskLo byte = (0xFF << bitAccelZOffuserPosLo)
)

// Used to convert bias values into a format that can be stored in the
// user registers.
const (
	accelOffuserMaxMg        int32 = 1000
	gyroOffuserMaxDps        int32 = 64
	gyroOffuserConfiguredDps int32 = 2000
)

////////////////////////////////////////////////////////////
/// COMMON FUNCTIONS
////////////////////////////////////////////////////////////

// Store the low 8 bits of cur_bias
func storeLowBits(cur_bias int16, offset byte) byte {
	return byte((cur_bias & 0x00FF) << int16(offset))
}

// Store the high 4 bits of cur_bias
func storeHighBits(cur_bias int16, offset byte) byte {
	return byte(((cur_bias & 0x0F00) >> 8) << int16(offset))
}

////////////////////////////////////////////////////////////
/// GYRO CALIBRATION
////////////////////////////////////////////////////////////

// Initialize gyro with sensible FSR, ODR, and filter values.
// Not clear if this is necessary, but it doesn't hurt.
func (i *IIM42652) initializeGyroForCalibration() error {
	// set the ODR to 1khz and FSR to 2000dps (but I think it's already set to this)
	gyroConfig, err := i.ReadRegister(RegisterGyroscopeConfig0)
	if err != nil {
		return err
	}

	gyroConfig &= ^bitGyroConfig0FSSelectMask
	gyroConfig |= bitGyroFSSelect2000dps

	gyroConfig &= ^bitGyroConfig0ODRMask
	gyroConfig |= bitGyroODRSelect1KHz

	err = i.WriteRegister(RegisterGyroscopeConfig0, gyroConfig)
	if err != nil {
		return err
	}

	// Set the Order filter
	gyroConfigBank1, err := i.ReadRegister(RegisterGyroscopeConfig1)
	if err != nil {
		return err
	}
	gyroConfigBank1 &= ^bitGyroConfig1GyroUiFiltOrdMask
	gyroConfigBank1 |= bitGyroConfig1GyroUiFilt3rdOrder
	err = i.WriteRegister(RegisterGyroscopeConfig1, gyroConfigBank1)
	if err != nil {
		return err
	}

	// Set the Bandwidth filter
	gyroAccelConfig, err := i.ReadRegister(RegisterAccelGyroConfig)
	if err != nil {
		return err
	}
	gyroAccelConfig &= ^bitGyroAccelConfig0GyroFiltMask
	gyroAccelConfig |= bitGyroUiFiltBandwidth
	err = i.WriteRegister(RegisterAccelGyroConfig, gyroAccelConfig)
	if err != nil {
		return err
	}

	time.Sleep(60 * time.Millisecond)
	return nil
}

func (i *IIM42652) AverageGyroSensorOutput(maxSamples int32) (average [3]int32, err error) {
	numSamples := int32(0)
	samplesDiscarded := int32(0)
	sum := [3]int32{0, 0, 0}
	for {
		gyroscopeData, err := i.GetGyroscopeData()
		if err != nil {
			return average, err
		}

		if gyroscopeData.RawX != -32768 && gyroscopeData.RawY != -32768 && gyroscopeData.RawZ != -32768 {
			sum[0] += int32(gyroscopeData.RawX)
			sum[1] += int32(gyroscopeData.RawY)
			sum[2] += int32(gyroscopeData.RawZ)
		} else {
			samplesDiscarded++
		}

		numSamples++
		time.Sleep(time.Millisecond)
		if numSamples == maxSamples {
			break
		}
	}

	// compute average value
	numSamples -= samplesDiscarded
	average[0] = (sum[0] / numSamples)
	average[1] = (sum[1] / numSamples)
	average[2] = (sum[2] / numSamples)
	return average, nil
}

// Negate the bias value and onvert the reading to the 64dps equivalent.
func convertGyroBiasToRegisterFormat(gyroBias int32) int16 {
	return int16(-(gyroBias * gyroOffuserConfiguredDps / gyroOffuserMaxDps) >> 4)
}

func (i *IIM42652) writeGyroBiasToUserRegister(bias [3]int32) error {
	// The 3 bias values are stored as 12 bits each.
	// They need to be stored interleaved across 5 byte registers
	// Look at the IIM42652 datasheet for more info.
	data := [5]byte{0, 0, 0, 0, 0}

	// The accelerometer bias data shares a register with the gyrosocope.
	// Copy the overlapping data to data[4] so that we don't lose it.
	accelData, err := i.ReadRegister(RegisterOffsetUser4)
	if err != nil {
		return err
	}
	data[4] = (accelData & bitAccelXOffuserMaskHi)

	cur_bias := convertGyroBiasToRegisterFormat(bias[0])
	data[0] = storeLowBits(cur_bias, bitGyroXOffuserPosLo)
	data[1] = storeHighBits(cur_bias, bitGyroXOffuserPosHi)

	cur_bias = convertGyroBiasToRegisterFormat(bias[1])
	data[1] |= storeHighBits(cur_bias, bitGyroYOffuserPosHi)
	data[2] = storeLowBits(cur_bias, bitGyroYOffuserPosLo)

	cur_bias = convertGyroBiasToRegisterFormat(bias[2])
	data[3] = storeLowBits(cur_bias, bitGyroZOffuserPosLo)
	data[4] |= storeHighBits(cur_bias, bitGyroZOffuserPosHi)

	userRegister := *RegisterOffsetUser0
	for idx := 0; idx < len(data); idx++ {
		err := i.WriteRegister(&userRegister, data[idx])
		if err != nil {
			return err
		}
		userRegister.Address += 1
	}

	return nil
}

// Calibrates the gyro by taking an average and
// storing the offset
func (i *IIM42652) CalibrateGyro(maxSamples int32) (bias [3]int32, err error) {
	err = i.initializeGyroForCalibration()
	if err != nil {
		return bias, err
	}
	time.Sleep(60 * time.Millisecond)

	bias, err = i.AverageGyroSensorOutput(maxSamples)
	if err != nil {
		return bias, err
	}

	err = i.writeGyroBiasToUserRegister(bias)
	if err != nil {
		return bias, err
	}

	return bias, nil
}

func (i *IIM42652) ClearGyroBias() error {
	return i.writeGyroBiasToUserRegister([3]int32{0, 0, 0})
}

////////////////////////////////////////////////////////////
/// ACCELEROMETER CALIBRATION
////////////////////////////////////////////////////////////

func (i *IIM42652) AverageAccelerometerSensorOutput(maxSamples int32) (average [3]int32, err error) {
	numSamples := int32(0)
	samplesDiscarded := int32(0)
	sum := [3]int32{0, 0, 0}
	for {
		accelerometerData, err := i.GetAcceleration()
		if err != nil {
			return average, err
		}

		if accelerometerData.RawX != -32768 && accelerometerData.RawY != -32768 && accelerometerData.RawZ != -32768 {
			sum[0] += int32(accelerometerData.RawX)
			sum[1] += int32(accelerometerData.RawY)
			sum[2] += int32(accelerometerData.RawZ)
		} else {
			samplesDiscarded++
		}

		numSamples++
		time.Sleep(time.Millisecond)
		if numSamples == maxSamples {
			break
		}
	}

	// compute average value
	numSamples -= samplesDiscarded
	average[0] = (sum[0] / numSamples)
	average[1] = (sum[1] / numSamples)
	average[2] = (sum[2] / numSamples)
	return average, nil
}

// Negate the bias value and convert the reading to the
// 1g sensitivity equivalent.
func convertAccelBiasToRegisterFormat(accelBias int32) int16 {
	return int16(-accelBias)
}

func (i *IIM42652) writeAccelerometerBiasToUserRegister(bias [3]int32) error {
	// The 3 bias values are stored as 12 bits each.
	// They need to be stored interleaved across 5 byte registers
	// Look at the IIM42652 datasheet for more info.
	data := [5]byte{0, 0, 0, 0, 0}

	// The gyroscope bias data shares a register with the accelerometer.
	// Copy the overlapping data to data[0] so that we don't lose it.
	gyroData, err := i.ReadRegister(RegisterOffsetUser4)
	if err != nil {
		return err
	}
	data[0] = (gyroData & bitGyroZOffuserMaskHi)

	cur_bias := convertAccelBiasToRegisterFormat(bias[0])
	data[0] |= storeHighBits(cur_bias, bitAccelXOffuserPosHi)
	data[1] = storeLowBits(cur_bias, bitAccelXOffuserPosLo)

	cur_bias = convertAccelBiasToRegisterFormat(bias[1])
	data[2] = storeLowBits(cur_bias, bitAccelYOffuserPosLo)
	data[3] = storeHighBits(cur_bias, bitAccelYOffuserPosHi)

	cur_bias = convertAccelBiasToRegisterFormat(bias[2])
	data[3] |= storeHighBits(cur_bias, bitAccelZOffuserPosHi)
	data[4] = storeLowBits(cur_bias, bitAccelZOffuserPosLo)

	userRegister := *RegisterOffsetUser4
	for idx := 0; idx < len(data); idx++ {
		err := i.WriteRegister(&userRegister, data[idx])
		if err != nil {
			return err
		}
		userRegister.Address += 1
	}

	return nil
}

func (i *IIM42652) CalibrateAccelerometer(maxSamples int32) (bias [3]int32, err error) {
	bias, err = i.AverageAccelerometerSensorOutput(maxSamples)
	if err != nil {
		return bias, err
	}

	err = i.writeAccelerometerBiasToUserRegister(bias)
	if err != nil {
		return bias, err
	}

	return bias, nil
}

func (i *IIM42652) ClearAccelerometerBias() error {
	return i.writeAccelerometerBiasToUserRegister([3]int32{0, 0, 0})
}
