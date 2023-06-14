package main

import (
	"fmt"
	"os"
	"time"

	geojson "github.com/paulmach/go.geojson"
	"github.com/spf13/cobra"
	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/direction"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/hivemapper-data-logger/data/sql"
	"github.com/streamingfast/hivemapper-data-logger/logger"
)

var ReplayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay a drive saved in an sql db",
	RunE:  replayE,
}

func init() {
	//IMU
	ReplayCmd.Flags().String("imu-config-file", "imu-logger.json", "imu logger config file")

	//DB
	ReplayCmd.Flags().String("db-import-path", "/mnt/data/gnss.v1.1.0.db", "path to sqliteLogger database")
	ReplayCmd.Flags().String("db-output-path", "/mnt/data/output.db", "path to sqliteLogger database")
	ReplayCmd.Flags().Duration("db-log-ttl", 12*time.Hour, "ttl of logs in database")
	ReplayCmd.Flags().BoolP("clean", "c", false, "purges output db where db-output-path is located before running rerun command")

	RootCmd.AddCommand(ReplayCmd)
}

func replayE(cmd *cobra.Command, _ []string) error {
	dbOutputPath := mustGetString(cmd, "db-output-path")
	clean := mustGetBool(cmd, "clean")
	if clean {
		err := os.Remove(dbOutputPath)
		if err != nil {
			return fmt.Errorf("failed to remove %s: %w", dbOutputPath, err)
		}
		fmt.Printf("Removed %s\n", dbOutputPath)
	}

	sqliteImporter := logger.NewSqlite(mustGetString(cmd, "db-import-path"), nil, nil)
	err := sqliteImporter.Init(0)
	if err != nil {
		return fmt.Errorf("initializing sqlite logger database: %w", err)
	}

	conf := imu.LoadConfig(mustGetString(cmd, "imu-config-file"))
	fmt.Println("Config: ", conf.String())

	dataHandler, err := NewDataHandler(
		mustGetString(cmd, "db-output-path"),
		mustGetDuration(cmd, "db-log-ttl"),
	)
	if err != nil {
		return fmt.Errorf("creating data handler: %w", err)
	}

	geoJsonHandler := NewGeoJsonHandler()

	directionEventFeed := direction.NewDirectionEventFeed(conf, dataHandler.HandleDirectionEvent, geoJsonHandler.HandleDirectionEvent)
	orientedEventFeed := imu.NewOrientedAccelerationFeed(directionEventFeed.HandleOrientedAcceleration, dataHandler.HandleOrientedAcceleration)
	tiltCorrectedAccelerationEventFeed := imu.NewTiltCorrectedAccelerationFeed(orientedEventFeed.HandleTiltCorrectedAcceleration)

	sqlFeed := sql.NewSqlImporterFeed(
		sqliteImporter,
		[]imu.RawFeedHandler{tiltCorrectedAccelerationEventFeed.HandleRawFeed, dataHandler.HandleRawImuFeed},
		[]gnss.GnssDataHandler{dataHandler.HandlerGnssData, directionEventFeed.HandleGnssData, geoJsonHandler.HandlerGnssData},
	)

	sqlFeed.Run()

	gj, err := geoJsonHandler.featureCollection.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshalling geojson: %w", err)
	}

	err = os.WriteFile("geo.json", gj, 0644)
	if err != nil {
		return fmt.Errorf("writing geojson: %w", err)
	}

	return nil
}

type GeoJsonHandler struct {
	geometry          *geojson.Geometry
	featureCollection *geojson.FeatureCollection
}

func NewGeoJsonHandler() *GeoJsonHandler {
	return &GeoJsonHandler{
		featureCollection: geojson.NewFeatureCollection(),
	}
}

func (h *GeoJsonHandler) HandlerGnssData(data *neom9n.Data) error {
	//fmt.Println("Got GnssData")
	h.geometry = geojson.NewPointGeometry([]float64{data.Longitude, data.Latitude})
	return nil
}
func (h *GeoJsonHandler) HandleDirectionEvent(e data.Event) error {
	//fmt.Println("Got DirectionEvent")
	feature := geojson.NewFeature(h.geometry)
	feature.Type = e.GetName()
	feature.SetProperty("event", e.GetName())
	h.featureCollection.AddFeature(feature)

	return nil
}
