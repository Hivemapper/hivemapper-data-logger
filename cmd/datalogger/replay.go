package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data"
	"github.com/Hivemapper/hivemapper-data-logger/data/direction"
	"github.com/Hivemapper/hivemapper-data-logger/data/gnss"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/Hivemapper/hivemapper-data-logger/data/sql"
	"github.com/Hivemapper/hivemapper-data-logger/logger"
	geojson "github.com/paulmach/go.geojson"
	"github.com/rosshemsley/kalman"
	"github.com/rosshemsley/kalman/models"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ReplayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay a drive saved in an sql db",
	RunE:  replayE,
}

func init() {
	//IMU
	ReplayCmd.Flags().String("imu-config-file", "imu-logger.json", "imu logger config file")
	ReplayCmd.Flags().String("imu-json-destination-folder", "imu", "json destination folder")
	ReplayCmd.Flags().Duration("imu-json-save-interval", 15*time.Second, "json save interval")
	ReplayCmd.Flags().String("imu-axis-map", "CamX:Z,CamY:X,CamZ:Y", "axis mapping of camera x,y,z values to real world x,y,z values. Default value are HDC mappings")
	ReplayCmd.Flags().String("imu-inverted", "X:false,Y:false,Z:false", "axis inverted mapping of x,y,z values")

	//GNSS
	ReplayCmd.Flags().String("gnss-json-destination-folder", "gps", "json destination folder")
	ReplayCmd.Flags().Duration("gnss-json-save-interval", 15*time.Second, "json save interval")

	//DB
	ReplayCmd.Flags().String("db-import-path", "gnss.v1.1.0.db", "path to sqliteLogger database")
	ReplayCmd.Flags().String("db-output-path", "output.db", "path to sqliteLogger database")
	ReplayCmd.Flags().Duration("db-log-ttl", 12*time.Hour, "ttl of logs in database")
	ReplayCmd.Flags().BoolP("clean", "c", false, "purges output db where db-output-path is located before running replay command")

	RootCmd.AddCommand(ReplayCmd)
}

func replayE(cmd *cobra.Command, _ []string) error {
	axisMap, err := parseAxisMap(mustGetString(cmd, "imu-axis-map"))
	if err != nil {
		return fmt.Errorf("parsing axis map: %w", err)
	}

	invX, invY, invZ, err := parseInvertedMappings(mustGetString(cmd, "imu-inverted"))
	if err != nil {
		return fmt.Errorf("parsing inverted mappings: %w", err)
	}

	axisMap.SetInvertedAxes(invX, invY, invZ)

	dbOutputPath := mustGetString(cmd, "db-output-path")
	clean := mustGetBool(cmd, "clean")
	if clean {
		_, err := os.Stat(dbOutputPath)
		if os.IsNotExist(err) {
			fmt.Println("No output db found, nothing to clean")
		} else {
			err := os.Remove(dbOutputPath)
			if err != nil {
				return fmt.Errorf("failed to remove %s: %w", dbOutputPath, err)
			}
			fmt.Printf("Removed %s\n", dbOutputPath)
		}
	}

	sqliteImporter := logger.NewSqlite(mustGetString(cmd, "db-import-path"), nil, nil, nil)
	err = sqliteImporter.Init(0)
	if err != nil {
		return fmt.Errorf("initializing sqlite logger database: %w", err)
	}

	conf := imu.LoadConfig(mustGetString(cmd, "imu-config-file"))
	fmt.Println("Config: ", conf.String())

	//todo: init data handler
	//todo: clear all folders
	//todo: copy database file
	//todo: make replay accessible from debugger
	//todo: from debugger stop file purger && restart odc-api ...
	//todo: emit and empty image for each imu event when filename change ...

	dataHandler, err := NewDataHandler(
		mustGetString(cmd, "db-output-path"),
		mustGetDuration(cmd, "db-log-ttl"),
		mustGetString(cmd, "gnss-json-destination-folder"),
		mustGetDuration(cmd, "gnss-json-save-interval"),
		mustGetString(cmd, "imu-json-destination-folder"),
		mustGetDuration(cmd, "imu-json-save-interval"),
		mustGetBool(cmd, "json-logs-enabled"),
		mustGetBool(cmd, "enable-redis-logs"),
		mustGetInt(cmd, "max-redis-imu-entries"),
		mustGetInt(cmd, "max-redis-mag-entries"),
		mustGetInt(cmd, "max-redis-gnss-entries"),
		mustGetInt(cmd, "max-redis-gnss-auth-entries"),
		getBoolOrDefault(cmd, "redis-log-pbtxt"),
	)
	if err != nil {
		return fmt.Errorf("creating data handler: %w", err)
	}

	geoJsonHandler := NewGeoJsonHandler()

	directionEventFeed := direction.NewDirectionEventFeed(conf,
		dataHandler.HandleDirectionEvent,
		geoJsonHandler.HandleDirectionEvent,
	)
	orientedEventFeed := imu.NewOrientedAccelerationFeed(
		directionEventFeed.HandleOrientedAcceleration,
		dataHandler.HandleOrientedAcceleration,
	)
	tiltCorrectedAccelerationEventFeed := imu.NewTiltCorrectedAccelerationFeed(orientedEventFeed.HandleTiltCorrectedAcceleration)

	sqlFeed := sql.NewSqlImporterFeed(
		sqliteImporter,
		[]imu.RawFeedHandler{
			tiltCorrectedAccelerationEventFeed.HandleRawFeed,
			dataHandler.HandleRawImuFeed,
		},
		[]gnss.GnssDataHandler{
			dataHandler.HandlerGnssData,
			directionEventFeed.HandleGnssData,
			geoJsonHandler.HandleGnss,
		},
	)

	err = sqlFeed.Run(axisMap)
	if err != nil {
		return fmt.Errorf("running sql feed: %w", err)
	}

	if len(geoJsonHandler.locationCollection.Features) > 0 {
		locations, err := geoJsonHandler.locationCollection.MarshalJSON()
		if err != nil {
			return fmt.Errorf("marshalling locations: %w", err)
		}

		err = os.WriteFile("locations.json", locations, 0644)
		if err != nil {
			return fmt.Errorf("writing locations: %w", err)
		}
	}

	if len(geoJsonHandler.fixedLocationCollection.Features) > 0 {
		fixedLocations, err := geoJsonHandler.fixedLocationCollection.MarshalJSON()
		if err != nil {
			return fmt.Errorf("marshalling geojson: %w", err)
		}

		err = os.WriteFile("fixed-locations.json", fixedLocations, 0644)
		if err != nil {
			return fmt.Errorf("writing geojson: %w", err)
		}
	}

	return nil
}

type GeoJsonHandler struct {
	geometry                *geojson.Geometry
	fixGeometry             *geojson.Geometry
	locationCollection      *geojson.FeatureCollection
	fixedLocationCollection *geojson.FeatureCollection
	initialized             bool

	lonModel      *models.SimpleModel
	lonFilter     *kalman.KalmanFilter
	latModel      *models.SimpleModel
	latFilter     *kalman.KalmanFilter
	headingModel  *models.SimpleModel
	headingFilter *kalman.KalmanFilter
	//mongoClient   *mongo.Client
	db              *mongo.Database
	pointCollection *mongo.Collection
}

func NewGeoJsonHandler() *GeoJsonHandler {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	db := client.Database("geo")
	pointCollection := db.Collection("points")

	return &GeoJsonHandler{
		db:                      db,
		pointCollection:         pointCollection,
		locationCollection:      geojson.NewFeatureCollection(),
		fixedLocationCollection: geojson.NewFeatureCollection(),
	}
}

func (h *GeoJsonHandler) init(now time.Time) {
	h.initialized = true

	h.lonModel = models.NewSimpleModel(now, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	h.lonFilter = kalman.NewKalmanFilter(h.lonModel)

	h.latModel = models.NewSimpleModel(now, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	h.latFilter = kalman.NewKalmanFilter(h.latModel)

	h.headingModel = models.NewSimpleModel(now, 0.0, models.SimpleModelConfig{
		InitialVariance:     0.0,
		ProcessVariance:     2.0,
		ObservationVariance: 2.0,
	})
	h.headingFilter = kalman.NewKalmanFilter(h.headingModel)
}

func (h *GeoJsonHandler) Update(lon, lat, heading float64, now time.Time) (float64, float64, float64, error) {
	err := h.lonFilter.Update(now, h.lonModel.NewMeasurement(lon))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("updating lon filter: %w", err)
	}
	newLon := h.lonModel.Value(h.lonFilter.State())

	err = h.latFilter.Update(now, h.latModel.NewMeasurement(lat))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("updating lat filter: %w", err)
	}
	newLat := h.latModel.Value(h.latFilter.State())

	err = h.headingFilter.Update(now, h.headingModel.NewMeasurement(heading))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("updating heading filter: %w", err)
	}

	return newLon, newLat, heading, nil
}

var lastGnssTime = time.Time{}
var lastEvent = ""

func (h *GeoJsonHandler) HandleGnss(data *neom9n.Data) error {
	if !h.initialized {
		h.init(data.Timestamp)
	}

	if data.Fix == "none" {
		return nil
	}

	if data.Timestamp != lastGnssTime {
		h.geometry = geojson.NewPointGeometry([]float64{data.Longitude, data.Latitude})
		lastGnssTime = data.Timestamp
		feature := geojson.NewFeature(h.geometry)
		feature.Type = "gnss"
		feature.SetProperty("event", lastEvent)
		feature.SetProperty("dop", data.Dop)
		feature.SetProperty("horizontalAccuracy", data.HorizontalAccuracy)
		feature.SetProperty("satellites", data.Satellites)
		feature.SetProperty("eph", data.Eph)
		feature.SetProperty("heading", data.Heading)
		feature.SetProperty("headingAccuracy", data.HeadingAccuracy)
		h.locationCollection.AddFeature(feature)

		//fixing localtion
		nLon, nLat, _, w1, w2, err := h.magic(data.Longitude, data.Latitude, data.Heading, data.Timestamp)
		if err != nil {
			return fmt.Errorf("magic: %w", err)
		}
		//fmt.Println("magic", nLon, nLat)

		h.fixGeometry = geojson.NewPointGeometry([]float64{nLon, nLat})
		lastGnssTime = data.Timestamp
		feature = geojson.NewFeature(h.fixGeometry)
		feature.Type = "gnss"
		feature.SetProperty("event", lastEvent)
		feature.SetProperty("origin", []float64{data.Longitude, data.Latitude})
		if w1 != nil {
			feature.SetProperty("w1", []float64{w1.Lon, w1.Lat})
		}
		if w2 != nil {
			feature.SetProperty("w2", []float64{w2.Lon, w2.Lat})
		}

		h.fixedLocationCollection.AddFeature(feature)
	}

	return nil
}

func (h *GeoJsonHandler) HandleDirectionEvent(e data.Event) error {
	if e.GetGnssData().Fix == "none" {
		return nil
	}

	eventName := e.GetName()
	if strings.Contains(eventName, "DETECTED") {
		lastEvent = e.GetName()
	} else {
		lastEvent = ""
	}

	return nil
}
