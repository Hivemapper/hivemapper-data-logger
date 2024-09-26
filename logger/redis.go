package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

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
	ttl time.Duration
}

func NewRedis() *Redis {
	return &Redis{}
}

func (s *Redis) Init(logTTL time.Duration) error {
	fmt.Println("Initializing Redis logger")
	s.DB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Use your Redis server address
	})
	fmt.Println("Getting context")
	s.ctx = context.Background()
	s.ttl = logTTL
	fmt.Println("Redis logger initialized")
	return nil
}

func (s *Redis) LogImuData(imudata []byte) error {
	// Push the JSON data to the Redis list
	if err := s.DB.LPush(s.ctx, "accelerometer_data", imudata, s.ttl).Err(); err != nil {
		return err
	}
	return nil
}
