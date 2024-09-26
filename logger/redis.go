package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/go-redis/redis/v8"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

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
	// lock sync.Mutex
	// DB                       *sql.DB
	// file                     string
	// doInsert                 bool
	// purgeQueryFuncList       []PurgeQueryFunc
	// createTableQueryFuncList []CreateTableQueryFunc
	// alterTableQueryFuncList []AlterTableQueryFunc

	// logs chan Sqlable

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
	jsondata, err := json.Marshal(imudata)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "accelerometer_data", jsondata).Err(); err != nil {
		return err
	}

	if err := s.DB.LTrim(s.ctx, "accelerometer_data", 0, 1000).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogMagnetometerData(magdata MagnetometerWrapper) error {
	// Serialize to json
	jsondata, err := json.Marshal(magdata)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "magnetometer_data", jsondata).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "magnetometer_data", 0, 1000).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogGnssData(gnssdata neom9n.Data) error {
	// Serialize to json
	jsondata, err := json.Marshal(gnssdata)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "gnss_data", jsondata).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "gnss_data", 0, 1000).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Redis) LogGnssAuthData(gnssAuthData neom9n.Data) error {
	// Serialize to json
	jsondata, err := json.Marshal(gnssAuthData)
	if err != nil {
		return err
	}

	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "gnss_auth_data", jsondata).Err(); err != nil {
		return err
	}
	if err := s.DB.LTrim(s.ctx, "gnss_auth_data", 0, 1000).Err(); err != nil {
		return err
	}
	return nil
}
