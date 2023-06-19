package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/direction"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/hivemapper-data-logger/data/merged"
	"github.com/streamingfast/hivemapper-data-logger/gen/proto/sf/events/v1/eventsv1connect"
	"github.com/streamingfast/hivemapper-data-logger/logger"
	"github.com/streamingfast/hivemapper-data-logger/webconnect"
	"github.com/streamingfast/imu-controller/device/iim42652"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var LogCmd = &cobra.Command{
	Use:   "log",
	Short: "Start the data logger",
	RunE:  logRun,
}

func init() {
	// Imu
	LogCmd.Flags().String("imu-config-file", "imu-logger.json", "Imu logger config file. Default path is ./imu-logger.json")

	// Gnss
	LogCmd.Flags().String("gnss-config-file", "gnss-logger.json", "Neom9n logger config file. Default path is ./gnss-logger.json")
	LogCmd.Flags().String("gnss-json-destination-folder", "/mnt/data/gps", "json destination folder")
	LogCmd.Flags().Duration("gnss-json-save-interval", 15*time.Second, "json save interval")
	LogCmd.Flags().String("gnss-serial-config-name", "/dev/ttyAMA1", "Config serial location")
	LogCmd.Flags().String("gnss-mga-offline-file-path", "/mnt/data/mgaoffline.ubx", "path to mga offline files")

	// Sqlite database
	LogCmd.Flags().String("db-output-path", "/mnt/data/gnss.v1.1.0.db", "path to sqliteLogger database")
	LogCmd.Flags().Duration("db-log-ttl", 12*time.Hour, "ttl of logs in database")

	// Connect-go
	LogCmd.Flags().String("listen-addr", ":9000", "address to listen on")

	RootCmd.AddCommand(LogCmd)
}

func logRun(cmd *cobra.Command, _ []string) error {
	imuDevice := iim42652.NewSpi("/dev/spidev0.0", iim42652.AccelerationSensitivityG16, iim42652.GyroScalesG2000, true)
	err := imuDevice.Init()
	if err != nil {
		return fmt.Errorf("initializing IMU: %w", err)
	}
	err = imuDevice.UpdateRegister(iim42652.RegisterAccelConfig, func(currentValue byte) byte {
		return currentValue | 0x01
	})
	if err != nil {
		return fmt.Errorf("failed to update register: %w", err)
	}
	conf := imu.LoadConfig(mustGetString(cmd, "imu-config-file"))
	fmt.Println("Config: ", conf.String())

	serialConfigName := mustGetString(cmd, "gnss-serial-config-name")
	mgaOfflineFilePath := mustGetString(cmd, "gnss-mga-offline-file-path")
	gnssDevice := neom9n.NewNeom9n(serialConfigName, mgaOfflineFilePath)
	err = gnssDevice.Init(nil)
	if err != nil {
		return fmt.Errorf("initializing neom9n: %w", err)
	}

	//TODO: write gnss data to json file
	listenAddr := mustGetString(cmd, "listen-addr")
	eventServer := webconnect.NewEventServer()

	dataHandler, err := NewDataHandler(
		mustGetString(cmd, "db-output-path"),
		mustGetDuration(cmd, "db-log-ttl"),
		mustGetString(cmd, "gnss-json-destination-folder"),
		mustGetDuration(cmd, "gnss-json-save-interval"),
	)
	if err != nil {
		return fmt.Errorf("creating data handler: %w", err)
	}

	directionEventFeed := direction.NewDirectionEventFeed(conf, dataHandler.HandleDirectionEvent, eventServer.HandleDirectionEvent)
	orientedEventFeed := imu.NewOrientedAccelerationFeed(directionEventFeed.HandleOrientedAcceleration, dataHandler.HandleOrientedAcceleration)
	tiltCorrectedAccelerationEventFeed := imu.NewTiltCorrectedAccelerationFeed(orientedEventFeed.HandleTiltCorrectedAcceleration)

	rawImuEventFeed := imu.NewRawFeed(imuDevice, tiltCorrectedAccelerationEventFeed.HandleRawFeed, dataHandler.HandleRawImuFeed)
	go func() {
		err := rawImuEventFeed.Run()
		if err != nil {
			panic(fmt.Errorf("running raw imu event feed: %w", err))
		}
	}()

	gnssEventFeed := gnss.NewGnssFeed(
		[]gnss.GnssDataHandler{
			dataHandler.HandlerGnssData,
			directionEventFeed.HandleGnssData,
		},
		nil,
	)
	go func() {
		err = gnssEventFeed.Run(gnssDevice)
		if err != nil {
			panic(fmt.Errorf("running gnss event feed: %w", err))
		}
	}()

	mux := http.NewServeMux()
	path, handler := eventsv1connect.NewEventServiceHandler(eventServer)

	opts := cors.Options{
		AllowedHeaders: []string{"*"},
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}
	handler = cors.New(opts).Handler(handler)

	mux.Handle(path, handler)

	fmt.Printf("Starting server on %s ...\n", listenAddr)
	err = http.ListenAndServe(listenAddr, h2c.NewHandler(mux, &http2.Server{}))

	if err != nil {
		return fmt.Errorf("running server: %w", err)
	}

	return nil
}

type DataHandler struct {
	sqliteLogger   *logger.Sqlite
	gnssJsonLogger *logger.JsonFile
	imuJsonLogger  *logger.JsonFile
	gnssData       *neom9n.Data
}

func NewDataHandler(dbPath string, dbLogTTL time.Duration, gnssJsonDestFolder string, gnssSaveInterval time.Duration) (*DataHandler, error) {
	sqliteLogger := logger.NewSqlite(
		dbPath,
		[]logger.CreateTableQueryFunc{merged.CreateTableQuery, merged.ImuRawCreateTableQuery, direction.CreateTableQuery},
		[]logger.PurgeQueryFunc{merged.PurgeQuery, merged.ImuRawPurgeQuery, direction.PurgeQuery})
	err := sqliteLogger.Init(dbLogTTL)
	if err != nil {
		return nil, fmt.Errorf("initializing sqlite logger database: %w", err)
	}

	gnssJsonLogger := logger.NewJsonFile(gnssJsonDestFolder, gnssSaveInterval)
	err = gnssJsonLogger.Init()
	if err != nil {
		return nil, fmt.Errorf("initializing gnss json logger: %w", err)
	}

	return &DataHandler{
		sqliteLogger:   sqliteLogger,
		gnssJsonLogger: gnssJsonLogger,
	}, err
}

func (h *DataHandler) HandleOrientedAcceleration(acceleration *imu.Acceleration, tiltAngles *imu.TiltAngles, orientation imu.Orientation) error {
	gnssData := mustGnssEvent(h.gnssData)
	err := h.sqliteLogger.Log(merged.NewSqlWrapper(acceleration, tiltAngles, orientation, gnssData))
	if err != nil {
		return fmt.Errorf("logging merged data to sqlite: %w", err)
	}
	return nil
}

func (h *DataHandler) HandlerGnssData(data *neom9n.Data) error {
	h.gnssData = data
	err := h.gnssJsonLogger.Log(data.Timestamp, data)
	if err != nil {
		return fmt.Errorf("logging gnss data to json: %w", err)
	}
	return nil
}

func (h *DataHandler) HandleRawImuFeed(acceleration *imu.Acceleration, angularRate *iim42652.AngularRate) error {
	gnssData := mustGnssEvent(h.gnssData)
	err := h.sqliteLogger.Log(merged.NewImuRawSqlWrapper(acceleration, gnssData))
	if err != nil {
		return fmt.Errorf("logging raw imu data to sqlite: %w", err)
	}
	return nil
}

func (h *DataHandler) HandleDirectionEvent(event data.Event) error {
	gnssData := mustGnssEvent(h.gnssData)
	err := h.sqliteLogger.Log(direction.NewSqlWrapper(event, gnssData))
	if err != nil {
		return fmt.Errorf("logging direction data to sqlite: %w", err)
	}
	return nil
}

func mustGnssEvent(e *neom9n.Data) *neom9n.Data {
	if e == nil {
		return &neom9n.Data{
			Dop:        &neom9n.Dop{},
			RF:         &neom9n.RF{},
			Satellites: &neom9n.Satellites{},
		}
	}

	return e
}
