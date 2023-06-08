package main

import (
	"fmt"
	"os"
	"time"

	geojson "github.com/paulmach/go.geojson"
	"github.com/spf13/cobra"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/hivemapper-data-logger/data/merged"
	"github.com/streamingfast/hivemapper-data-logger/data/sql"
	"github.com/streamingfast/hivemapper-data-logger/logger"
	"github.com/streamingfast/shutter"
)

var ReRunCmd = &cobra.Command{
	Use:   "rerun",
	Short: "Rerun the sql logger",
	RunE:  reRunE,
}

func init() {
	//IMU
	ReRunCmd.Flags().String("imu-config-file", "imu-logger.json", "imu logger config file")

	//DB
	ReRunCmd.Flags().String("db-import-path", "/mnt/data/gnss.v1.1.0.db", "path to sqliteLogger database")
	ReRunCmd.Flags().String("db-output-path", "/mnt/data/output.db", "path to sqliteLogger database")
	ReRunCmd.Flags().Duration("db-log-ttl", 12*time.Hour, "ttl of logs in database")

	// todo: add remove out.db option

	RootCmd.AddCommand(ReRunCmd)
}

func reRunE(cmd *cobra.Command, _ []string) error {
	sh := shutter.New()
	sqliteOutput := logger.NewSqlite(
		mustGetString(cmd, "db-output-path"),
		[]logger.CreateTableQueryFunc{merged.CreateTableQuery, imu.CreateTableQuery},
		nil,
	)
	err := sqliteOutput.Init(0)
	if err != nil {
		return fmt.Errorf("initializing sqlite logger database: %w", err)
	}

	sqliteImporter := logger.NewSqlite(mustGetString(cmd, "db-import-path"), nil, nil)
	err = sqliteImporter.Init(0)
	if err != nil {
		return fmt.Errorf("initializing sqlite logger database: %w", err)
	}
	sqlFeed := sql.NewSqlFeed(sqliteImporter)
	sqlFeed.OnTerminated(sh.Shutdown)

	conf := imu.LoadConfig(mustGetString(cmd, "imu-config-file"))
	fmt.Println("Config: ", conf.String())

	correctedImuEventFeed := imu.NewCorrectedAccelerationFeed()
	correctedImuEventFeed.Start(sqlFeed.SubscribeImu("imu-corrected"))

	directionEventFeed := imu.NewDirectionEventFeed(conf)
	directionEventFeed.Start(correctedImuEventFeed.Subscribe("imu-direction"))
	mergedEventFeed := data.NewEventFeedMerger(
		sqlFeed.SubscribeImu("merger"),
		sqlFeed.SubscribeGnss("merger"),
		correctedImuEventFeed.Subscribe("merger"),
		directionEventFeed.Subscribe("merger"),
	)

	mergedEventFeed.Start()
	mergedEventSub := mergedEventFeed.Subscribe("rerun")
	sqlFeed.Start()

	fmt.Println("Sql import started")
	fmt.Println("Waiting for events...")

	var imuRawEvent *imu.RawImuEvent
	var correctedImuEvent *imu.CorrectedAccelerationEvent
	var gnssEvent *gnss.GnssEvent

	featureCollection := geojson.NewFeatureCollection()
	var geometry *geojson.Geometry

	for !sh.IsTerminating() && !sh.IsTerminated() {
		select {
		case <-sh.Terminating():
			fmt.Println("Terminating")
			break
		case e := <-mergedEventSub.IncomingEvents:
			switch e := e.(type) {
			case *imu.RawImuEvent:
				imuRawEvent = e
			case *imu.CorrectedAccelerationEvent:
				correctedImuEvent = e
			case *gnss.GnssEvent:
				gnssEvent = e
				geometry = geojson.NewPointGeometry([]float64{e.Data.Latitude, e.Data.Longitude})
			}
			if e.GetCategory() == "DIRECTION_CHANGE" {
				err := sqliteOutput.Log(imu.NewSqlWrapper(e, mustGnssEvent(gnssEvent)))
				if err != nil {
					return fmt.Errorf("logging to sqlite: %w", err)
				}

				feature := geojson.NewFeature(geometry)
				feature.Type = e.GetName()
				feature.SetProperty("event", e.GetName())
				featureCollection.AddFeature(feature)
			}
		}
		if imuRawEvent != nil && correctedImuEvent != nil {
			ge := mustGnssEvent(gnssEvent)
			w := merged.NewSqlWrapper(imuRawEvent, correctedImuEvent, ge)
			err = sqliteOutput.Log(w)
			if err != nil {
				return fmt.Errorf("logging to sqlite: %w", err)
			}
			imuRawEvent = nil
			correctedImuEvent = nil
		}
	}
	gj, err := featureCollection.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshalling geojson: %w", err)
	}

	err = os.WriteFile("geo.json", gj, 0644)
	if err != nil {
		return fmt.Errorf("writing geojson: %w", err)
	}

	return nil
}