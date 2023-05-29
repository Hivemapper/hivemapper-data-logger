package iim42652

import (
	"encoding/hex"
	"fmt"
)

const ShortMax = 32767
const PowerOnSleep = 250
const ReadMask byte = 0x80

type Bank byte

const (
	Bank0 Bank = 0x00
	Bank1 Bank = 0x01
	Bank2 Bank = 0x02
	Bank4 Bank = 0x04
)

func (b Bank) ToBytes() []byte {
	return []byte{byte(b)}
}

func (b Bank) String() string {
	return hex.EncodeToString(b.ToBytes())
}

type Address byte

func (r Address) ToBytes() []byte {
	return []byte{byte(r)}
}

func (r Address) String() string {
	return hex.EncodeToString(r.ToBytes())
}

type Register struct {
	Bank    Bank
	Address Address
}

func (r *Register) String() string {
	return fmt.Sprintf("bank %s addr:%s", r.Bank, r.Address)
}

var (
	RegisterBankSel            = &Register{Bank0, 0x76}
	RegisterIntfConfig1        = &Register{Bank0, 0x4D}
	RegisterAccelDataX1        = &Register{Bank0, 0x1f}
	RegisterPwrMgmt0           = &Register{Bank0, 0x4e}
	RegisterAccelWomXThreshold = &Register{Bank4, 0x4a}
	RegisterAccelWomYThreshold = &Register{Bank4, 0x4b}
	RegisterAccelWomZThreshold = &Register{Bank4, 0x4c}
	RegisterIntSource1         = &Register{Bank0, 0x66}
	RegisterIntSource4         = &Register{Bank0, 0x69}
	RegisterSdmConfig0         = &Register{Bank0, 0x57}
	RegisterIntStatus2         = &Register{Bank0, 0x37}
	RegisterAccelConfig        = &Register{Bank0, 0x50}
)

const (
	//RegisterBankSel      Address = 0x76
	DevCfgReg    Address = 0x11
	TempData1Reg Address = 0x1d
	//RegisterAccelDataX1  Address = 0x1f
	AccelDataX0Reg Address = 0x20
	AccelDataY1Reg Address = 0x21
	AccelDataY0Reg Address = 0x22
	AccelDataZ1Reg Address = 0x23
	AccelDataZ0Reg Address = 0x24
	GyroDataX1Reg  Address = 0x25
	//RegisterPwrMgmt0     Address = 0x4e
	GyroConfig0Reg  Address = 0x4f
	AccelConfig0Reg Address = 0x50
)

const (
	GyroModeLowNoise          byte = 0x0c
	AccelerometerModeLowPower byte = 0x02
	AccelerometerModeLowNoise byte = 0x03
)

const (
	ConfigScaleMask  byte  = 0x07
	ConfigScaleShift uint8 = 5
	ConfigRateMask   byte  = 0x0f
)

//const (
//	TempDataSize  uint8 = 2
//	AccelDataSize uint8 = 6
//	GyroDataSize  uint8 = 6
//)

const (
	Dps2000 uint16 = iota
	Dps1000
	Dps500
	Dps250
	Dps125
	Dps62_5
	Dps31_25
	Dps15_625
)
