package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	storage "github.com/Hivemapper/hivemapper-data-logger/proto-out"
	"github.com/go-redis/redis/v8"
	"github.com/streamingfast/imu-controller/device/iim42652"
	"google.golang.org/protobuf/encoding/prototext"
)

const UseJson = false

type MagnetometerWrapper struct {
	System_time time.Time `json:"system_time"`
	Mag_x       float64   `json:"mag_x"`
	Mag_y       float64   `json:"mag_y"`
	Mag_z       float64   `json:"mag_z"`
}

type ImuDataWrapper2 struct {
	System_time time.Time `json:"system_time"`
	Accel       *Accel    `json:"accel"`
	Gyro        *Gyro     `json:"gyro"`
	Temp        float64   `json:"temp"`
	Time        time.Time `json:"time"`
}

func NewImuDataWrapper2(system_time time.Time, temperature iim42652.Temperature, acceleration *imu.Acceleration, angularRate *iim42652.AngularRate) *ImuDataWrapper2 {
	return &ImuDataWrapper2{
		System_time: system_time,
		Accel:       NewAccel(acceleration.X, acceleration.Y, acceleration.Z),
		Gyro:        NewGyro(angularRate.X, angularRate.Y, angularRate.Z),
		Time:        acceleration.Time,
		Temp:        *temperature,
	}
}

func NewMagnetometerWrapper(system_time time.Time, mag_x float64, mag_y float64, mag_z float64) *MagnetometerWrapper {
	return &MagnetometerWrapper{
		System_time: system_time,
		Mag_x:       mag_x,
		Mag_y:       mag_y,
		Mag_z:       mag_z,
	}
}

type Redis struct {
	DB  *redis.Client
	ctx context.Context
}

func NewRedis() *Redis {
	return &Redis{}
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

func (s *Redis) LogImuData(imudata ImuDataWrapper2) error {

	// Serialize to json

	var data []byte = nil
	if UseJson {
		jsondata, err := json.Marshal(imudata)
		data = jsondata
		if err != nil {
			return err
		}
	} else {
		// create imu proto
		newdata := storage.ImuData{
			SystemTime: imudata.System_time.String(),
			Accelerometer: &storage.ImuData_AccelerometerData{
				X: imudata.Accel.X,
				Y: imudata.Accel.Y,
				Z: imudata.Accel.Z,
			},
			Gyroscope: &storage.ImuData_GyroscopeData{
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
		data = protodata
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "imu_data", data).Err(); err != nil {
		return err
	}

	if err := s.DB.LTrim(s.ctx, "imu_data", 0, 1000).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogMagnetometerData(magdata MagnetometerWrapper) error {

	var data []byte = nil
	if UseJson {
		// Serialize to json
		jsondata, err := json.Marshal(magdata)
		data = jsondata
		if err != nil {
			return err
		}
	} else {
		// create magnetometer proto
		newdata := storage.MagnetometerData{
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
		data = protodata
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "magnetometer_data", data).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "magnetometer_data", 0, 1000).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogGnssData(gnssdata neom9n.Data) error {
	var data []byte = nil
	if UseJson {
		// Serialize to json
		jsondata, err := json.Marshal(gnssdata)
		data = jsondata
		if err != nil {
			return err
		}
	} else {
		// Create gnss proto
		newdata := storage.GnssData{
			Ttff:      gnssdata.Ttff,
			Fix:       gnssdata.Fix,
			Latitude:  gnssdata.Latitude,
			Longitude: gnssdata.Longitude,
			Altitude:  gnssdata.Altitude,
			Speed:     gnssdata.Speed,
		}
		// serialize the data
		protodata, err := prototext.Marshal(&newdata)
		if err != nil {
			return err
		}
		data = protodata
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "gnss_data", data).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "gnss_data", 0, 1000).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogGnssAuthData(gnssAuthData neom9n.Data) error {
	var data []byte = nil
	if UseJson {
		// Serialize to json
		jsondata, err := json.Marshal(gnssAuthData)
		data = jsondata
		if err != nil {
			return err
		}
	} else {
		// Use protobuf
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "gnss_auth_data", data).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "gnss_auth_data", 0, 1000).Err(); err != nil {
		return err
	}
	return nil
}
