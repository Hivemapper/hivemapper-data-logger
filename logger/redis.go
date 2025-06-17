package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	sensordata "github.com/Hivemapper/hivemapper-data-logger/proto-out"
	"github.com/daedaleanai/ublox/ubx"
	"github.com/go-redis/redis/v8"
	"github.com/streamingfast/imu-controller/device/iim42652"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
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
	logProtoText       bool
}

func NewRedis(maxImuEntries int, maxMagEntries int, maxGnssEntries int, maxGnssAuthEntries int, logProtoText bool) *Redis {
	return &Redis{
		maxImuEntries:      maxImuEntries,
		maxMagEntries:      maxMagEntries,
		maxGnssEntries:     maxGnssEntries,
		maxGnssAuthEntries: maxGnssAuthEntries,
		logProtoText:       logProtoText,
	}
}

func (s *Redis) Init() error {
	fmt.Println("Initializing Redis logger")
	s.DB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	s.ctx = context.Background()

	pong, err := s.DB.Ping(s.ctx).Result()
	if err != nil {
		fmt.Printf("Could not connect to Redis: %v\n", err)
		return fmt.Errorf("ping pong failed")
	}
	fmt.Println("Redis connected:", pong)
	return nil
}

func (s *Redis) Marshal(message proto.Message) ([]byte, error) {
	if s.logProtoText {
		return prototext.Marshal(message)
	}
	return proto.Marshal(message)
}

func (s *Redis) LogImuData(imudata ImuRedisWrapper) error {
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
	protodata, err := s.Marshal(&newdata)
	if err != nil {
		return err
	}
	if err := s.DB.LPush(s.ctx, "imu_data", protodata).Err(); err != nil {
		return err
	}
	return s.DB.LTrim(s.ctx, "imu_data", 0, int64(s.maxImuEntries)).Err()
}

func (s *Redis) LogImuDataBatch(batch []*ImuRedisWrapper) error {
	if len(batch) == 0 {
		return nil
	}
	pipe := s.DB.Pipeline()
	for _, imudata := range batch {
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
		protodata, err := s.Marshal(&newdata)
		if err != nil {
			continue
		}
		pipe.LPush(s.ctx, "imu_data", protodata)
	}
	pipe.LTrim(s.ctx, "imu_data", 0, int64(s.maxImuEntries))
	_, err := pipe.Exec(s.ctx)
	return err
}

func (s *Redis) LogMagnetometerDataBatch(batch []*MagnetometerRedisWrapper) error {
	if len(batch) == 0 {
		return nil
	}
	pipe := s.DB.Pipeline()
	for _, mag := range batch {
		newdata := sensordata.MagnetometerData{
			SystemTime: mag.System_time.String(),
			X:          mag.Mag_x,
			Y:          mag.Mag_y,
			Z:          mag.Mag_z,
		}
		protodata, err := s.Marshal(&newdata)
		if err != nil {
			continue
		}
		pipe.LPush(s.ctx, "magnetometer_data", protodata)
	}
	pipe.LTrim(s.ctx, "magnetometer_data", 0, int64(s.maxMagEntries))
	_, err := pipe.Exec(s.ctx)
	return err
}

func (s *Redis) LogGnssDataBatch(batch []*neom9n.Data) error {
	if len(batch) == 0 {
		return nil
	}
	pipe := s.DB.Pipeline()
	for _, gnss := range batch {
		newdata := sensordata.GnssData{
			Ttff:                gnss.Ttff,
			SystemTime:          gnss.SystemTime.String(),
			ActualSystemTime:    gnss.ActualSystemTime.String(),
			Timestamp:           gnss.Timestamp.String(),
			Fix:                 gnss.Fix,
			Latitude:            gnss.Latitude,
			UnfilteredLatitude:  gnss.UnfilteredLatitude,
			Longitude:           gnss.Longitude,
			UnfilteredLongitude: gnss.UnfilteredLongitude,
			Altitude:            gnss.Altitude,
			Heading:             gnss.Heading,
			Speed:               gnss.Speed,
			Dop: &sensordata.GnssData_Dop{
				Hdop: gnss.Dop.HDop,
				Vdop: gnss.Dop.VDop,
				Tdop: gnss.Dop.TDop,
				Gdop: gnss.Dop.GDop,
				Pdop: gnss.Dop.PDop,
				Xdop: gnss.Dop.XDop,
				Ydop: gnss.Dop.YDop,
			},
			Satellites: &sensordata.GnssData_Satellites{
				Seen: int64(gnss.Satellites.Seen),
				Used: int64(gnss.Satellites.Used),
			},
			Sep:               gnss.Sep,
			Eph:               gnss.Eph,
			SpeedAccuracy:     gnss.SpeedAccuracy,
			HeadingAccuracy:   gnss.HeadingAccuracy,
			TimeResolved:      int32(gnss.TimeResolved),
			HorizontalAccuracy: gnss.HorizontalAccuracy,
			VerticalAccuracy:   gnss.VerticalAccuracy,
			Gga:                gnss.GGA,
			Cno:                gnss.Cno,
			PosConfidence:      gnss.PosConfidence,
			Rf: &sensordata.GnssData_RF{
				JammingState: gnss.RF.JammingState,
				AntStatus:    gnss.RF.AntStatus,
				AntPower:     gnss.RF.AntPower,
				PostStatus:   gnss.RF.PostStatus,
				NoisePerMs:   uint32(gnss.RF.NoisePerMS),
				AgcCnt:       uint32(gnss.RF.AgcCnt),
				JamInd:       uint32(gnss.RF.JamInd),
				OfsI:         int32(gnss.RF.OfsI),
				MagI:         int32(gnss.RF.MagI),
				OfsQ:         int32(gnss.RF.OfsQ),
				MagQ:         int32(gnss.RF.MagQ),
			},
		}
		if gnss.RxmMeasx != nil {
			newdata.RxmMeasx = &sensordata.GnssData_RxmMeasx{
				GpsTowMs: gnss.RxmMeasx.GpsTOW_ms,
				GloTowMs: gnss.RxmMeasx.GloTOW_ms,
				BdsTowMs: gnss.RxmMeasx.BdsTOW_ms,
			}
		}
		protodata, err := s.Marshal(&newdata)
		if err != nil {
			continue
		}
		pipe.LPush(s.ctx, "gnss_data", protodata)
	}
	pipe.LTrim(s.ctx, "gnss_data", 0, int64(s.maxGnssEntries))
	_, err := pipe.Exec(s.ctx)
	return err
}
