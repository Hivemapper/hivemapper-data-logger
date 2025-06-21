package logger

import (
	"time"

	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

type Accel struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func NewAccel(x, y, z float64) *Accel {
	return &Accel{
		X: x,
		Y: y,
		Z: z,
	}
}

type Gyro struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func NewGyro(x, y, z float64) *Gyro {
	return &Gyro{
		X: x,
		Y: y,
		Z: z,
	}
}

type Fsync struct {
	TimeDelta int16 `json:"time_delta"`
	FsyncInt  bool  `json:"fsync_int"`
}

func NewFsync(time_delta int16, fsyncInt bool) *Fsync {
	return &Fsync{
		TimeDelta: time_delta,
		FsyncInt:  fsyncInt,
	}
}

type ImuDataWrapper struct {
	Accel *Accel    `json:"accel"`
	Gyro  *Gyro     `json:"gyro"`
	Temp  float64   `json:"temp"`
	Time  time.Time `json:"time"`
	Fsync *Fsync    `json:"fsync"`
}

func NewImuDataWrapper(temperature iim42652.Temperature, acceleration *imu.Acceleration, angularRate *iim42652.AngularRate, fsync *iim42652.Fsync) *ImuDataWrapper {
	return &ImuDataWrapper{
		Accel: NewAccel(acceleration.X, acceleration.Y, acceleration.Z),
		Gyro:  NewGyro(angularRate.X, angularRate.Y, angularRate.Z),
		Time:  acceleration.Time,
		Temp:  *temperature,
		Fsync: NewFsync(fsync.TimeDelta, fsync.FsyncInt),
	}
}
