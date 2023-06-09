package sql

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/streamingfast/shutter"

	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/hivemapper-data-logger/logger"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

var LIMIT = 200

type SqlFeed struct {
	*shutter.Shutter
	sqlite            *logger.Sqlite
	imuSubscriptions  data.Subscriptions
	gnssSubscriptions data.Subscriptions
}

func NewSqlFeed(sqlite *logger.Sqlite) *SqlFeed {
	return &SqlFeed{
		Shutter:           shutter.New(),
		sqlite:            sqlite,
		imuSubscriptions:  make(data.Subscriptions),
		gnssSubscriptions: make(data.Subscriptions),
	}
}

func (s *SqlFeed) SubscribeImu(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	s.imuSubscriptions[name] = sub
	return sub
}

func (s *SqlFeed) SubscribeGnss(name string) *data.Subscription {
	sub := &data.Subscription{
		IncomingEvents: make(chan data.Event),
	}
	s.gnssSubscriptions[name] = sub
	return sub
}

func (s *SqlFeed) Start() {
	fmt.Println("Starting sql feed")

	numOfRows := 0
	err := s.sqlite.SingleRowQuery("select count(*) from merged;", func(rows *sql.Rows) error {
		err := rows.Scan(&numOfRows)
		if err != nil {
			return fmt.Errorf("scanning last position: %s", err.Error())
		}
		return nil
	}, nil)
	fmt.Printf("found rows %d in merged table\n", numOfRows)
	if err != nil {
		panic(fmt.Errorf("failed to fetch number of rows in table merged: %w", err))
	}

	go func() {
		numOfIterations := int(math.Floor(float64(numOfRows/LIMIT)) + 1)

		lastGnssSystemTime := time.Time{}
		for i := 0; i < numOfIterations; i++ {
			offset := LIMIT * i
			fmt.Printf("Fetching events from sqlite database with offset %d\n", offset)
			err := s.sqlite.Query(false, query(offset), func(rows *sql.Rows) error {
				acceleration := &iim42652.Acceleration{}
				gnssData := &neom9n.Data{
					SystemTime: time.Time{},
					Timestamp:  time.Time{},
					Dop:        &neom9n.Dop{},
					Satellites: &neom9n.Satellites{},
					RF:         &neom9n.RF{},
				}
				err := rows.Scan(
					&acceleration.TotalMagnitude,
					&acceleration.Z, // CamX -> Z
					&acceleration.X, // CamY -> X
					&acceleration.Y, // CamZ -> Y
					&gnssData.SystemTime,
					&gnssData.Timestamp,
					&gnssData.Fix,
					&gnssData.Ttff,
					&gnssData.Latitude,
					&gnssData.Longitude,
					&gnssData.Altitude,
					&gnssData.Speed,
					&gnssData.Heading,
					&gnssData.Satellites.Seen,
					&gnssData.Satellites.Used,
					&gnssData.Eph,
					&gnssData.HorizontalAccuracy,
					&gnssData.VerticalAccuracy,
					&gnssData.HeadingAccuracy,
					&gnssData.SpeedAccuracy,
					&gnssData.Dop.HDop,
					&gnssData.Dop.VDop,
					&gnssData.Dop.XDop,
					&gnssData.Dop.YDop,
					&gnssData.Dop.TDop,
					&gnssData.Dop.PDop,
					&gnssData.Dop.GDop,
					&gnssData.RF.JammingState,
					&gnssData.RF.AntStatus,
					&gnssData.RF.AntPower,
					&gnssData.RF.PostStatus,
					&gnssData.RF.NoisePerMS,
					&gnssData.RF.AgcCnt,
					&gnssData.RF.JamInd,
					&gnssData.RF.OfsI,
					&gnssData.RF.MagI,
					&gnssData.RF.OfsQ,
				)
				if err != nil {
					return fmt.Errorf("scanning last position: %s", err.Error())
				}

				rawImuEvent := imu.NewRawImuEvent(acceleration, nil)
				gnssEvent := gnss.NewGnssEvent(gnssData)

				for _, sub := range s.imuSubscriptions {
					sub.IncomingEvents <- rawImuEvent
				}

				if gnssData.SystemTime != lastGnssSystemTime {
					lastGnssSystemTime = gnssData.SystemTime
					for _, sub := range s.gnssSubscriptions {
						sub.IncomingEvents <- gnssEvent
					}
				}

				return nil
			}, nil)

			if err != nil {
				panic(fmt.Errorf("failed to query database: %w", err))
			}
		}
		fmt.Println("Finished sql feed")
		s.Shutdown(nil)
	}()
}

func query(offset int) string {
	return fmt.Sprintf(`
		select imu_total_magnitude,
			   imu_acc_x,
			   imu_acc_y,
			   imu_acc_z,
			   gnss_system_time,
			   gnss_time,
			   gnss_fix,
			   gnss_ttff,
			   gnss_latitude,
			   gnss_longitude,
			   gnss_altitude,
			   gnss_speed,
			   gnss_heading,
			   gnss_satellites_seen,
			   gnss_satellites_used,
			   gnss_eph,
			   gnss_horizontal_accuracy,
			   gnss_vertical_accuracy,
			   gnss_heading_accuracy,
			   gnss_speed_accuracy,
			   gnss_dop_h,
			   gnss_dop_v,
			   gnss_dop_x,
			   gnss_dop_y,
			   gnss_dop_t,
			   gnss_dop_p,
			   gnss_dop_g,
			   gnss_rf_jamming_state,
			   gnss_rf_ant_status,
			   gnss_rf_ant_power,
			   gnss_rf_post_status,
			   gnss_rf_noise_per_ms,
			   gnss_rf_agc_cnt,
			   gnss_rf_jam_ind,
			   gnss_rf_ofs_i,
			   gnss_rf_mag_i,
			   gnss_rf_ofs_q
		from merged order by id asc limit %d offset %d;
		`, LIMIT, offset,
	)
}
