package sql

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/hivemapper-data-logger/logger"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

var LIMIT = 200

type SqlImporterFeed struct {
	sqlite              *logger.Sqlite
	imuRawFeedHandlers  []imu.RawFeedHandler
	gssDataFeedHandlers []gnss.GnssDataHandler
}

func NewSqlImporterFeed(sqlite *logger.Sqlite, imuRawFeedHandlers []imu.RawFeedHandler, gssDataFeedHandlers []gnss.GnssDataHandler) *SqlImporterFeed {
	return &SqlImporterFeed{
		sqlite:              sqlite,
		imuRawFeedHandlers:  imuRawFeedHandlers,
		gssDataFeedHandlers: gssDataFeedHandlers,
	}
}

func (s *SqlImporterFeed) Run() {
	fmt.Println("Starting sql feed")

	numOfRows := 0
	err := s.sqlite.SingleRowQuery("select count(*) from imu_raw;", func(rows *sql.Rows) error {
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

	numOfIterations := int(math.Floor(float64(numOfRows/LIMIT)) + 1)

	lastGnssSystemTime := time.Time{}
	for i := 0; i < numOfIterations; i++ {
		offset := LIMIT * i
		err := s.sqlite.Query(false, query(offset), func(rows *sql.Rows) error {
			acceleration := &imu.Acceleration{}
			gnssData := &neom9n.Data{
				SystemTime: time.Time{},
				Timestamp:  time.Time{},
				Dop:        &neom9n.Dop{},
				Satellites: &neom9n.Satellites{},
				RF:         &neom9n.RF{},
			}
			err := rows.Scan(
				&acceleration.Time,
				&acceleration.X,
				&acceleration.Y,
				&acceleration.Z,
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

			ar := &iim42652.AngularRate{}
			for _, handler := range s.imuRawFeedHandlers {
				m := math.Sqrt(acceleration.X*acceleration.X + acceleration.Y*acceleration.Y + acceleration.Z*acceleration.Z)
				acceleration.Magnitude = m
				err := handler(acceleration, ar)
				if err != nil {
					return fmt.Errorf("failed to handle imu raw feed: %w", err)
				}
			}

			if gnssData.SystemTime != lastGnssSystemTime {
				lastGnssSystemTime = gnssData.SystemTime
				for _, handler := range s.gssDataFeedHandlers {
					err := handler(gnssData)
					if err != nil {
						return fmt.Errorf("failed to handle gnss data feed: %w", err)
					}
				}
			}
			return nil
		}, nil)

		if err != nil {
			panic(fmt.Errorf("failed to query database: %w", err))
		}
	}
	fmt.Println("Finished sql feed")

}

func query(offset int) string {
	return fmt.Sprintf(`
		select 
               imu_time,
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
