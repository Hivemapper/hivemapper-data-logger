package main

import (
	"fmt"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"net/http"
	"time"

	"github.com/rs/cors"
	"github.com/streamingfast/hivemapper-data-logger/gen/proto/sf/events/v1/eventsv1connect"
	"github.com/streamingfast/hivemapper-data-logger/webconnect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/spf13/cobra"
	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/hivemapper-data-logger/logger"
	"github.com/streamingfast/imu-controller/device/iim42652"
)

var LogCmd = &cobra.Command{
	Use:   "log",
	Short: "Start the data logger",
	RunE:  logRun,
}

func init() {
	// Imu
	LogCmd.Flags().String("imu-config-file", "imu-logger.json", "Imu logger config file. Default path is ./imu-logger.json")

	// GNSS
	LogCmd.Flags().String("gnss-config-file", "gnss-logger.json", "Neom9n logger config file. Default path is ./gnss-logger.json")
	LogCmd.Flags().String("gnss-json-destination-folder", "/mnt/data/gps", "json destination folder")
	LogCmd.Flags().Duration("gnss-json-save-interval", 15*time.Second, "json save interval")
	LogCmd.Flags().Int64("gnss-json-destination-folder-max-size", int64(30000*1024), "json destination folder maximum size") // 30MB
	LogCmd.Flags().String("gnss-serial-config-name", "/dev/ttyAMA1", "Config serial location")
	LogCmd.Flags().String("gnss-mga-offline-file-path", "/mnt/data/mgaoffline.ubx", "path to mga offline files")
	LogCmd.Flags().String("gnss-db-path", "/mnt/data/gnss.v1.0.3.db", "path to sqliteLogger database")
	LogCmd.Flags().Duration("gnss-db-log-ttl", 12*time.Hour, "ttl of logs in database")

	// Connect-go
	LogCmd.Flags().String("listen-addr", ":9000", "address to listen on")

	RootCmd.AddCommand(LogCmd)
}

func logRun(cmd *cobra.Command, args []string) error {
	imuDevice := iim42652.NewSpi("/dev/spidev0.0", iim42652.AccelerationSensitivityG16, iim42652.GyroScalesG2000, true)
	err := imuDevice.Init()
	if err != nil {
		return fmt.Errorf("initializing IMU: %w", err)
	}

	serialConfigName := mustGetString(cmd, "gnss-serial-config-name")
	mgaOfflineFilePath := mustGetString(cmd, "gnss-mga-offline-file-path")
	dbPath := mustGetString(cmd, "gnss-db-path")
	logTTl := mustGetDuration(cmd, "gnss-db-log-ttl")
	jsonDestinationFolder := mustGetString(cmd, "gnss-json-destination-folder")
	jsonSaveInterval := mustGetDuration(cmd, "gnss-json-save-interval")
	jsonDestinationFolderMaxSize := mustGetInt64(cmd, "gnss-json-destination-folder-max-size")

	gnssDevice := neom9n.NewNeom9n(serialConfigName, mgaOfflineFilePath)
	sqliteLogger := logger.NewSqlite(dbPath, gnss.CreateTableQuery, gnss.PurgeQuery)
	jsonLogger := logger.NewJsonFile(jsonDestinationFolder, jsonDestinationFolderMaxSize, jsonSaveInterval)

	err = imuDevice.UpdateRegister(iim42652.RegisterAccelConfig, func(currentValue byte) byte {
		return currentValue | 0x01
	})

	if err != nil {
		return fmt.Errorf("failed to update register: %w", err)
	}

	conf := imu.LoadConfig(mustGetString(cmd, "imu-config-file"))
	fmt.Println("Config: ", conf.String())

	imuEventFeed := imu.NewEventFeed(imuDevice, conf)
	go func() {
		err := imuEventFeed.Run()
		if err != nil {
			panic(fmt.Errorf("running pipeline: %w", err))
		}
	}()

	err = sqliteLogger.Init(logTTl)
	if err != nil {
		return fmt.Errorf("initializing sqlite database: %w", err)
	}
	lastPosition, err := gnss.GetLastPosition(sqliteLogger)
	if err != nil {
		return fmt.Errorf("getting last posotion: %w", err)
	}

	gnssEventFeed := gnss.NewEventFeed()
	gnssFileLoggerSubscription := gnssEventFeed.Subscribe("gnss-file-logger")
	err = jsonLogger.Init(gnssFileLoggerSubscription)
	if err != nil {
		return fmt.Errorf("initializing json logger database: %w", err)
	}

	err = gnssDevice.Init(lastPosition)
	if err != nil {
		return fmt.Errorf("initializing neom9n: %w", err)
	}
	dataFeed := neom9n.NewDataFeed(gnssEventFeed.HandleData)

	go func() {
		err = gnssDevice.Run(dataFeed, func(now time.Time) {
			dataFeed.SetStartTime(now)
			jsonLogger.StartStoring()
			sqliteLogger.StartStoring()
		})
		if err != nil {
			panic(fmt.Errorf("running gnss: %w", err))
		}
	}()
	go func() {
		sub := gnssEventFeed.Subscribe("gnss-sql-logger")
		for {
			select {
			case event := <-sub.IncomingEvents:
				e := event.(*gnss.GnssEvent)
				err = sqliteLogger.Log(gnss.NewSqlWrapper(e.Data))
				if err != nil {
					panic(fmt.Errorf("writing to file: %w", err))
				}
			}
		}
	}()

	//todo: init file logger for imu
	//todo: init db logger for imu

	grpcImuSubscription := imuEventFeed.Subscribe("grpc")
	grpcGnssSubscription := gnssEventFeed.Subscribe("grpc")

	//todo: feed merger that merge events from multiple feeds into one
	//todo: feed merger should offer a way to subscribe to a feed that is a merge of multiple feeds
	//todo: Move events filter to feed merger

	//todo: should like this .. eventServer := webconnect.NewEventServer(mergedEventFeed)

	merger := data.NewEventFeedMerger(grpcImuSubscription, grpcGnssSubscription)

	listenAddr := mustGetString(cmd, "listen-addr")
	eventServer := webconnect.NewEventServer(merger)
	mux := http.NewServeMux()
	path, handler := eventsv1connect.NewEventServiceHandler(eventServer)

	opts := cors.Options{
		AllowedHeaders: []string{"*"},
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}
	handler = cors.New(opts).Handler(handler)

	mux.Handle(path, handler)
	err = http.ListenAndServe(listenAddr, h2c.NewHandler(mux, &http2.Server{}))

	if err != nil {
		return fmt.Errorf("running server: %w", err)
	}

	return nil
}
