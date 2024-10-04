package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	sensordata "github.com/Hivemapper/hivemapper-data-logger/proto-out"
	"github.com/go-redis/redis/v8"
	"github.com/streamingfast/imu-controller/device/iim42652"
	"google.golang.org/protobuf/encoding/prototext"
)

type MagnetometerRedisWrapper struct {
	System_time time.Time `json:"system_time"`
	Mag_x       float64   `json:"mag_x"`
	Mag_y       float64   `json:"mag_y"`
	Mag_z       float64   `json:"mag_z"`
}

type ImuRedisWrapper struct {
	System_time time.Time `json:"system_time"`
	Accel       *Accel    `json:"accel"`
	Gyro        *Gyro     `json:"gyro"`
	Temp        float64   `json:"temp"`
	Time        time.Time `json:"time"`
}

func NewImuRedisWrapper(system_time time.Time, temperature iim42652.Temperature, acceleration *imu.Acceleration, angularRate *iim42652.AngularRate) *ImuRedisWrapper {
	return &ImuRedisWrapper{
		System_time: system_time,
		Accel:       NewAccel(acceleration.X, acceleration.Y, acceleration.Z),
		Gyro:        NewGyro(angularRate.X, angularRate.Y, angularRate.Z),
		Time:        acceleration.Time,
		Temp:        *temperature,
	}
}

func NewMagnetometerRedisWrapper(system_time time.Time, mag_x float64, mag_y float64, mag_z float64) *MagnetometerRedisWrapper {
	return &MagnetometerRedisWrapper{
		System_time: system_time,
		Mag_x:       mag_x,
		Mag_y:       mag_y,
		Mag_z:       mag_z,
	}
}

type Redis struct {
	DB                 *redis.Client
	ctx                context.Context
	maxImuEntries      int
	maxMagEntries      int
	maxGnssEntries     int
	maxGnssAuthEntries int
}

func NewRedis(maxImuEntries int, maxMagEntries int, maxGnssEntries int, maxGnssAuthEntries int) *Redis {
	return &Redis{
		maxImuEntries:      maxImuEntries,
		maxMagEntries:      maxMagEntries,
		maxGnssEntries:     maxGnssEntries,
		maxGnssAuthEntries: maxGnssAuthEntries,
	}
}

func (s *Redis) Init() error {
	fmt.Println("Initializing Redis logger")
	s.DB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Use your Redis server address
	})
	fmt.Println("Getting context")
	s.ctx = context.Background()

	// Test the connection with a PING command
	pong, err := s.DB.Ping(s.ctx).Result()
	if err != nil {
		fmt.Printf("Could not connect to Redis: %v\n", err)
		return fmt.Errorf("ping pong failed")
	}
	fmt.Println("Redis connected:", pong)
	fmt.Println("Redis logger initialized")
	return nil
}

func (s *Redis) LogImuData(imudata ImuRedisWrapper) error {
	// create imu proto
	newdata := sensordata.ImuData{
		SystemTime: imudata.System_time.String(),
		Accelerometer: &sensordata.ImuData_AccelerometerData{
			X: imudata.Accel.X,
			Y: imudata.Accel.Y,
			Z: imudata.Accel.Z,
		},
		Gyroscope: &sensordata.ImuData_GyroscopeData{
			X: imudata.Gyro.X,
			Y: imudata.Gyro.Y,
			Z: imudata.Gyro.Z,
		},
		Temperature: imudata.Temp,
		Time:        imudata.Time.String(),
	}
	// serialize the data
	protodata, err := prototext.Marshal(&newdata)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "imu_data", protodata).Err(); err != nil {
		return err
	}

	if err := s.DB.LTrim(s.ctx, "imu_data", 0, int64(s.maxImuEntries)).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogMagnetometerData(magdata MagnetometerRedisWrapper) error {
	// create magnetometer proto
	newdata := sensordata.MagnetometerData{
		SystemTime: magdata.System_time.String(),
		X:          magdata.Mag_x,
		Y:          magdata.Mag_y,
		Z:          magdata.Mag_z,
	}
	// serialize the data
	protodata, err := prototext.Marshal(&newdata)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "magnetometer_data", protodata).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "magnetometer_data", 0, int64(s.maxMagEntries)).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogGnssData(gnssdata neom9n.Data) error {
	// Create gnss proto
	newdata := sensordata.GnssData{
		Ttff:             gnssdata.Ttff,
		SystemTime:       gnssdata.SystemTime.String(),
		ActualSystemTime: gnssdata.ActualSystemTime.String(),
		Timestamp:        gnssdata.Timestamp.String(),
		Fix:              gnssdata.Fix,
		Latitude:         gnssdata.Latitude,
		Longitude:        gnssdata.Longitude,
		Altitude:         gnssdata.Altitude,
		Heading:          gnssdata.Heading,
		Speed:            gnssdata.Speed,
		Dop: &sensordata.GnssData_Dop{
			Hdop: gnssdata.Dop.HDop,
			Vdop: gnssdata.Dop.VDop,
			Tdop: gnssdata.Dop.TDop,
			Gdop: gnssdata.Dop.GDop,
			Pdop: gnssdata.Dop.PDop,
			Xdop: gnssdata.Dop.XDop,
			Ydop: gnssdata.Dop.YDop,
		},
		Satellites: &sensordata.GnssData_Satellites{
			Seen: int64(gnssdata.Satellites.Seen),
			Used: int64(gnssdata.Satellites.Used),
		},
		Sep: gnssdata.Sep,
		Eph: gnssdata.Eph,
		Rf: &sensordata.GnssData_RF{
			JammingState: gnssdata.RF.JammingState,
			AntStatus:    gnssdata.RF.AntStatus,
			AntPower:     gnssdata.RF.AntPower,
			PostStatus:   gnssdata.RF.PostStatus,
			NoisePerMs:   uint32(gnssdata.RF.NoisePerMS),
			AgcCnt:       uint32(gnssdata.RF.AgcCnt),
			JamInd:       uint32(gnssdata.RF.JamInd),
			OfsI:         int32(gnssdata.RF.OfsI),
			MagI:         int32(gnssdata.RF.MagI),
			OfsQ:         int32(gnssdata.RF.OfsQ),
			MagQ:         int32(gnssdata.RF.MagQ),
		},
		SpeedAccuracy:      gnssdata.SpeedAccuracy,
		HeadingAccuracy:    gnssdata.HeadingAccuracy,
		HorizontalAccuracy: gnssdata.HorizontalAccuracy,
		VerticalAccuracy:   gnssdata.VerticalAccuracy,
		Gga:                gnssdata.GGA,
		RxmMeasx: &sensordata.GnssData_RxmMeasx{
			GpsTowMs: gnssdata.RxmMeasx.GpsTOW_ms,
			GloTowMs: gnssdata.RxmMeasx.GloTOW_ms,
			BdsTowMs: gnssdata.RxmMeasx.BdsTOW_ms,
		},
	}
	// serialize the data
	protodata, err := prototext.Marshal(&newdata)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "gnss_data", protodata).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "gnss_data", 0, int64(s.maxGnssEntries)).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogGnssAuthData(gnssAuthData neom9n.Data) error {
	// Create gnss auth proto
	newdata := sensordata.GnssData{
		SystemTime: gnssAuthData.SystemTime.String(),
		SecEcsign: &sensordata.GnssData_UbxSecEcsign{
			Version:        uint32(gnssAuthData.SecEcsign.Version),
			MsgNum:         uint32(gnssAuthData.SecEcsign.MsgNum),
			FinalHash:      gnssAuthData.SecEcsign.FinalHash[:],
			SessionId:      gnssAuthData.SecEcsign.SessionId[:],
			EcdsaSignature: gnssAuthData.SecEcsign.EcdsaSignature[:],
		},
		SecEcsignBuffer: gnssAuthData.SecEcsignBuffer,
	}
	protodata, err := prototext.Marshal(&newdata)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "gnss_auth_data", protodata).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "gnss_auth_data", 0, int64(s.maxGnssAuthEntries)).Err(); err != nil {
		return err
	}
	return nil
}
