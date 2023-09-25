package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	gmux "github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data"
	"github.com/streamingfast/hivemapper-data-logger/data/direction"
	"github.com/streamingfast/hivemapper-data-logger/data/gnss"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/hivemapper-data-logger/data/merged"
	"github.com/streamingfast/hivemapper-data-logger/download"
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
	LogCmd.Flags().String("imu-json-destination-folder", "/mnt/data/imu", "json destination folder")
	LogCmd.Flags().Duration("imu-json-save-interval", 15*time.Second, "json save interval")
	LogCmd.Flags().String("imu-axis-map", "CamX:Z,CamY:X,CamZ:Y", "axis mapping of camera x,y,z values to real world x,y,z values. Default value is HDC mappings")
	LogCmd.Flags().String("imu-inverted", "X:false,Y:false,Z:false", "axis inverted mapping of x,y,z values")
	LogCmd.Flags().Bool("imu-skip-power-management", false, "skip power management setup of imu device on HDC-S")

	// Gnss
	LogCmd.Flags().Int("gnss-initial-baud-rate", 38400, "initial baud rate of gnss device")
	LogCmd.Flags().String("gnss-config-file", "gnss-logger.json", "Neom9n logger config file. Default path is ./gnss-logger.json")
	LogCmd.Flags().String("gnss-json-destination-folder", "/mnt/data/gps", "json destination folder")
	LogCmd.Flags().Duration("gnss-json-save-interval", 15*time.Second, "json save interval")
	LogCmd.Flags().String("gnss-dev-path", "/dev/ttyAMA1", "Config serial location")
	LogCmd.Flags().String("gnss-mga-offline-file-path", "/mnt/data/mgaoffline.ubx", "path to mga offline files")
	LogCmd.Flags().Bool("gnss-fix-check", true, "check if gnss fix is set")

	// Sqlite database
	LogCmd.Flags().String("db-output-path", "/mnt/data/gnss.v1.1.0.db", "path to sqliteLogger database")
	LogCmd.Flags().Duration("db-log-ttl", 12*time.Hour, "ttl of logs in database")
	LogCmd.Flags().String("imu-dev-path", "/dev/spidev0.0", "Config serial location")

	//Image feed
	LogCmd.Flags().String("images-folder", "/mnt/data/pic", "")

	// Connect-go
	LogCmd.Flags().String("listen-addr", ":9000", "address to listen on")

	// Http server
	LogCmd.Flags().String("http-listen-addr", ":9001", "http server address to listen on")

	LogCmd.Flags().Bool("skip-filtering", false, "skip filtering of gnss data")

	RootCmd.AddCommand(LogCmd)
}

func logRun(cmd *cobra.Command, _ []string) error {
	axisMap, err := parseAxisMap(mustGetString(cmd, "imu-axis-map"))
	if err != nil {
		return fmt.Errorf("parsing axis map: %w", err)
	}

	invX, invY, invZ, err := parseInvertedMappings(mustGetString(cmd, "imu-inverted"))
	if err != nil {
		return fmt.Errorf("parsing inverted mappings: %w", err)
	}

	axisMap.SetInvertedAxes(invX, invY, invZ)

	imuDevice := iim42652.NewSpi(
		mustGetString(cmd, "imu-dev-path"),
		iim42652.AccelerationSensitivityG16,
		iim42652.GyroScalesG2000,
		true,
		mustGetBool(cmd, "imu-skip-power-management"),
	)

	err = imuDevice.Init()
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

	serialConfigName := mustGetString(cmd, "gnss-dev-path")
	mgaOfflineFilePath := mustGetString(cmd, "gnss-mga-offline-file-path")
	gnssDevice := neom9n.NewNeom9n(serialConfigName, mgaOfflineFilePath, mustGetInt(cmd, "gnss-initial-baud-rate"))
	err = gnssDevice.Init(nil)
	if err != nil {
		return fmt.Errorf("initializing neom9n: %w", err)
	}

	listenAddr := mustGetString(cmd, "listen-addr")
	eventServer := webconnect.NewEventServer()

	dataHandler, err := NewDataHandler(
		mustGetString(cmd, "db-output-path"),
		mustGetDuration(cmd, "db-log-ttl"),
		mustGetString(cmd, "gnss-json-destination-folder"),
		mustGetDuration(cmd, "gnss-json-save-interval"),
		mustGetString(cmd, "imu-json-destination-folder"),
		mustGetDuration(cmd, "imu-json-save-interval"),
	)
	if err != nil {
		return fmt.Errorf("creating data handler: %w", err)
	}

	//directionEventFeed := direction.NewDirectionEventFeed(conf, dataHandler.HandleDirectionEvent, eventServer.HandleDirectionEvent)
	//orientedEventFeed := imu.NewOrientedAccelerationFeed(directionEventFeed.HandleOrientedAcceleration, dataHandler.HandleOrientedAcceleration)
	//tiltCorrectedAccelerationEventFeed := imu.NewTiltCorrectedAccelerationFeed(orientedEventFeed.HandleTiltCorrectedAcceleration)

	// TODO: implement replay image feed
	//imagesFeed := camera.NewImageFeed(mustGetString(cmd, "images-folder"), dataHandler.HandleImage)
	//go func() {
	//	err := imagesFeed.Run()
	//	if err != nil {
	//		panic(fmt.Errorf("running image feed: %w", err))
	//	}
	//}()

	rawImuEventFeed := imu.NewRawFeed(
		imuDevice,
		//tiltCorrectedAccelerationEventFeed.HandleRawFeed,
		dataHandler.HandleRawImuFeed,
	)
	go func() {
		err := rawImuEventFeed.Run(axisMap)
		if err != nil {
			panic(fmt.Errorf("running raw imu event feed: %w", err))
		}
	}()

	var options []gnss.Option
	if mustGetBool(cmd, "skip-filtering") {
		options = append(options, gnss.WithSkipFiltering())
	}
	gnssEventFeed := gnss.NewGnssFeed(
		[]gnss.GnssDataHandler{
			dataHandler.HandlerGnssData,
			//directionEventFeed.HandleGnssData,
			eventServer.HandleGnssData,
		},
		nil,
		options...,
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

	go func() {
		fmt.Printf("Starting GRPC server on %s ...\n", listenAddr)
		err = http.ListenAndServe(listenAddr, h2c.NewHandler(mux, &http2.Server{}))
		if err != nil {
			panic(fmt.Sprintf("running server: %s", err.Error()))
		}
	}()

	httpListenAddr := mustGetString(cmd, "http-listen-addr")

	origins := handlers.AllowedOrigins([]string{"*"})
	headers := handlers.AllowedHeaders([]string{"Content-Type"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	router := gmux.NewRouter().StrictSlash(true)

	down := download.NewDownload(dataHandler.sqliteLogger)
	router.HandleFunc("/rawData", down.GetRawData)
	router.HandleFunc("/debug/download", down.GetDatabaseFiles)

	err = http.ListenAndServe(httpListenAddr, handlers.CORS(origins, headers, methods)(router))
	fmt.Printf("Starting http server on %s ...\n", httpListenAddr)
	if err != nil {
		return fmt.Errorf("running http server: %w", err)
	}

	return nil
}

type DataHandler struct {
	sqliteLogger      *logger.Sqlite
	gnssJsonLogger    *logger.JsonFile
	imuJsonLogger     *logger.JsonFile
	gnssData          *neom9n.Data
	lastImageFileName string
}

func NewDataHandler(
	dbPath string,
	dbLogTTL time.Duration,
	gnssJsonDestFolder string,
	gnssSaveInterval time.Duration,
	imuJsonDestFolder string,
	imuSaveInterval time.Duration,
) (*DataHandler, error) {
	sqliteLogger := logger.NewSqlite(
		dbPath,
		[]logger.CreateTableQueryFunc{merged.CreateTableQuery, merged.ImuRawCreateTableQuery, direction.CreateTableQuery},
		[]logger.PurgeQueryFunc{merged.PurgeQuery, merged.ImuRawPurgeQuery, direction.PurgeQuery})
	err := sqliteLogger.Init(dbLogTTL)
	if err != nil {
		return nil, fmt.Errorf("initializing sqlite logger database: %w", err)
	}

	gnssJsonLogger := logger.NewJsonFile(gnssJsonDestFolder, gnssSaveInterval)
	err = gnssJsonLogger.Init(false)
	if err != nil {
		return nil, fmt.Errorf("initializing gnss json logger: %w", err)
	}

	imuJsonLogger := logger.NewJsonFile(imuJsonDestFolder, imuSaveInterval)
	err = imuJsonLogger.Init(true)
	if err != nil {
		return nil, fmt.Errorf("initializing imu json logger: %w", err)
	}

	return &DataHandler{
		sqliteLogger:   sqliteLogger,
		gnssJsonLogger: gnssJsonLogger,
		imuJsonLogger:  imuJsonLogger,
	}, err
}

func (h *DataHandler) HandleImage(imageFileName string) error {
	h.lastImageFileName = imageFileName
	return nil
}

func (h *DataHandler) HandleOrientedAcceleration(
	acceleration *imu.Acceleration,
	tiltAngles *imu.TiltAngles,
	temperature iim42652.Temperature,
	orientation imu.Orientation,
) error {
	gnssData := mustGnssEvent(h.gnssData)
	err := h.sqliteLogger.Log(merged.NewSqlWrapper(acceleration, tiltAngles, gnssData, temperature, orientation))
	if err != nil {
		return fmt.Errorf("logging merged data to sqlite: %w", err)
	}
	return nil
}

func (h *DataHandler) HandlerGnssData(data *neom9n.Data) error {
	h.gnssData = data
	if !h.gnssJsonLogger.IsLogging && data.Fix != "none" {
		h.gnssJsonLogger.StartStoring()
	}
	err := h.gnssJsonLogger.Log(data.Timestamp, data)

	if err != nil {
		return fmt.Errorf("logging gnss data to json: %w", err)
	}
	return nil
}

func (h *DataHandler) HandleRawImuFeed(acceleration *imu.Acceleration, angularRate *iim42652.AngularRate, temperature iim42652.Temperature) error {
	gnssData := mustGnssEvent(h.gnssData)
	err := h.sqliteLogger.Log(merged.NewImuRawSqlWrapper(temperature, acceleration, gnssData /*h.lastImageFileName*/))
	if err != nil {
		return fmt.Errorf("logging raw imu data to sqlite: %w", err)
	}
	imuDataWrapper := logger.NewImuDataWrapper(temperature, acceleration, angularRate)
	err = h.imuJsonLogger.Log(time.Now(), imuDataWrapper)
	if err != nil {
		return fmt.Errorf("logging raw imu data to json: %w", err)
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

func parseAxisMap(axisMapping string) (*iim42652.AxisMap, error) {
	if !strings.Contains(axisMapping, ",") {
		return nil, fmt.Errorf("axis mapping must contain ','")
	}

	if !strings.Contains(axisMapping, ":") {
		return nil, fmt.Errorf("axis mapping must contain ':'")
	}

	axes := strings.Split(axisMapping, ",")
	if len(axes) != 3 {
		return nil, fmt.Errorf("axis mapping must contain 3 axes")
	}

	xAxis := axes[0]
	if len(strings.Split(xAxis, ":")) != 2 {
		return nil, fmt.Errorf("x axis mapping must contain 2 parts separated by ':'")
	}

	yAxis := axes[1]
	if len(strings.Split(yAxis, ":")) != 2 {
		return nil, fmt.Errorf("y axis mapping must contain 2 parts separated by ':'")
	}
	zAxis := axes[2]
	if len(strings.Split(zAxis, ":")) != 2 {
		return nil, fmt.Errorf("z axis mapping must contain 2 parts separated by ':'")
	}

	return iim42652.NewAxisMap(
		strings.Split(xAxis, ":")[1],
		strings.Split(yAxis, ":")[1],
		strings.Split(zAxis, ":")[1],
	), nil
}

func parseInvertedMappings(invertedMapping string) (bool, bool, bool, error) {
	if !strings.Contains(invertedMapping, ",") {
		return false, false, false, fmt.Errorf("inverted mapping must contain ','")
	}

	if !strings.Contains(invertedMapping, ":") {
		return false, false, false, fmt.Errorf("inverted mapping must contain ':'")
	}

	axes := strings.Split(invertedMapping, ",")
	if len(axes) != 3 {
		return false, false, false, fmt.Errorf("inverted mapping must contain 3 axes")
	}

	xAxis := axes[0]
	if len(strings.Split(xAxis, ":")) != 2 {
		return false, false, false, fmt.Errorf("x inverted mapping must contain 2 parts separated by ':'")
	}

	yAxis := axes[1]
	if len(strings.Split(yAxis, ":")) != 2 {
		return false, false, false, fmt.Errorf("y inverted mapping must contain 2 parts separated by ':'")
	}

	zAxis := axes[2]
	if len(strings.Split(zAxis, ":")) != 2 {
		return false, false, false, fmt.Errorf("z inverted mapping must contain 2 parts separated by ':'")
	}

	invX, err := strconv.ParseBool(strings.Split(xAxis, ":")[1])
	if err != nil {
		return false, false, false, fmt.Errorf("parsing x inverted mapping: %w", err)
	}

	invY, err := strconv.ParseBool(strings.Split(yAxis, ":")[1])
	if err != nil {
		return false, false, false, fmt.Errorf("parsing y inverted mapping: %w", err)
	}

	invZ, err := strconv.ParseBool(strings.Split(zAxis, ":")[1])
	if err != nil {
		return false, false, false, fmt.Errorf("parsing z inverted mapping: %w", err)
	}

	return invX, invY, invZ, nil
}
