package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/streamingfast/gnss-controller/device/neom9n"

	geojson "github.com/paulmach/go.geojson"
	"github.com/spf13/cobra"
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

	//GNSS
	ReplayCmd.Flags().String("gnss-json-destination-folder", "/mnt/data/gps", "json destination folder")
	ReplayCmd.Flags().Duration("gnss-json-save-interval", 15*time.Second, "json save interval")

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
		mustGetString(cmd, "gnss-json-destination-folder"),
		mustGetDuration(cmd, "gnss-json-save-interval"),
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
		[]gnss.GnssDataHandler{dataHandler.HandlerGnssData, directionEventFeed.HandleGnssData, geoJsonHandler.HandleGnss},
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

	locations, err := geoJsonHandler.locationCollection.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshalling locations: %w", err)
	}

	err = os.WriteFile("locations.json", locations, 0644)
	if err != nil {
		return fmt.Errorf("writing locations: %w", err)
	}

	return nil
}

type GeoJsonHandler struct {
	geometry           *geojson.Geometry
	featureCollection  *geojson.FeatureCollection
	locationCollection *geojson.FeatureCollection
}

func NewGeoJsonHandler() *GeoJsonHandler {
	return &GeoJsonHandler{
		featureCollection:  geojson.NewFeatureCollection(),
		locationCollection: geojson.NewFeatureCollection(),
	}
}

var lastGnssTime = time.Time{}
var lastEvent = ""

func (h *GeoJsonHandler) HandleGnss(data *neom9n.Data) error {
	h.geometry = geojson.NewPointGeometry([]float64{data.Longitude, data.Latitude})
	if data.Timestamp != lastGnssTime {
		lastGnssTime = data.Timestamp
		feature := geojson.NewFeature(h.geometry)
		feature.Type = "gnss"
		feature.SetProperty("event", lastEvent)
		feature.SetProperty("dop", data.Dop)
		feature.SetProperty("horizontalAccuracy", data.HorizontalAccuracy)
		feature.SetProperty("satellites", data.Satellites)
		feature.SetProperty("eph", data.Eph)

		h.locationCollection.AddFeature(feature)
	}
	return nil
}

func (h *GeoJsonHandler) HandleDirectionEvent(e data.Event) error {
	h.geometry = geojson.NewPointGeometry([]float64{e.GetGnssData().Longitude, e.GetGnssData().Latitude})

	eventName := e.GetName()
	if strings.Contains(eventName, "DETECTED") {
		lastEvent = e.GetName()
	} else {
		lastEvent = ""
	}

	feature := geojson.NewFeature(h.geometry)
	feature.Type = e.GetName()
	feature.SetProperty("event", e.GetName())
	h.featureCollection.AddFeature(feature)

	return nil
}
