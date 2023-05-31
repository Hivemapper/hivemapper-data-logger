package neom9n

import (
	"errors"
	"fmt"
	"github.com/daedaleanai/ublox/ubx"
	"github.com/streamingfast/gnss-controller/logger"
	"github.com/streamingfast/gnss-controller/message"
	"github.com/streamingfast/gnss-controller/message/handlers"
	"log"
	"os"
	"time"

	"github.com/tarm/serial"
)

type JsonConfigFolder struct {
	destinationFolder string
	saveInterval      time.Duration
	maxFolderSize     int64
}

func NewJsonConfigFolder(destinationFolder string, saveInterval time.Duration, maxFolderSize int64) *JsonConfigFolder {
	return &JsonConfigFolder{
		destinationFolder: destinationFolder,
		saveInterval:      saveInterval,
		maxFolderSize:     maxFolderSize,
	}
}

type Neom9n struct {
	logTTL             time.Duration
	sqliteLogger       *logger.Sqlite
	jsonLogger         *logger.JsonFile
	loggerData         *logger.Data
	config             *serial.Config
	handlersRegistry   *message.HandlerRegistry
	mgaOfflineFilePath string
}

func NewNeom9n(logTTL time.Duration, databasePath string, jsonConfig *JsonConfigFolder, serialConfigName string, mgaOfflineFilePath string) *Neom9n {
	n := &Neom9n{
		logTTL:       logTTL,
		sqliteLogger: logger.NewSqlite(databasePath),
		jsonLogger:   logger.NewJsonFile(jsonConfig.destinationFolder, jsonConfig.maxFolderSize, jsonConfig.saveInterval),
		config: &serial.Config{
			Name:     serialConfigName, // /dev/ttyAMA1
			Baud:     38400,
			Parity:   serial.ParityNone,
			StopBits: serial.Stop1,
		},
		handlersRegistry:   message.NewHandlerRegistry(),
		mgaOfflineFilePath: mgaOfflineFilePath,
	}

	n.loggerData = logger.NewLoggerData(n.sqliteLogger, n.jsonLogger)

	return n
}

func (n *Neom9n) Init() {
	startTime := time.Now()

	err := n.sqliteLogger.Init(n.logTTL)
	handleError("initializing sqliteLogger", err)

	err = n.jsonLogger.Init()
	handleError("initializing json logger", err)

	n.loggerData.SetStartTime(startTime)
}

func (n *Neom9n) Run() {
	decoder := message.NewDecoder(n.handlersRegistry)

	stream, err := serial.OpenPort(n.config)
	handleError("opening gps serial port", err)

	output := make(chan ubx.Message)
	go func() {
		for {
			msg := <-output
			if _, ok := msg.(*ubx.MonRf); ok {
				encoded, err := ubx.EncodeReq(msg)
				_, err = stream.Write(encoded)
				handleError("writing message:", err)
			}
			encoded, err := ubx.Encode(msg)
			_, err = stream.Write(encoded)
			handleError("writing message:", err)
			//fmt.Printf("Sent: %T\n", msg)
			//fmt.Println("sent:", hex.EncodeToString(encoded))
		}
	}()

	cfg := ubx.CfgValSet{
		Version: 0x00,
		Layers:  ubx.CfgValSetLayers(ubx.CfgValSetLayersRam | ubx.CfgValSetLayersFlash | ubx.CfgValSetLayersBBR),
		CfgData: []*ubx.CfgData{
			{
				Key:   546374490, //0x2091035a CFG-MSGOUT-UBX_MON_RF_UART1
				Value: []byte{0x01},
			},
		},
	}

	fmt.Println("sending cfg val set")
	output <- &cfg

	lastPosition, err := n.sqliteLogger.GetLastPosition()
	if err != nil {
		handleError("getting last position from sqliteLogger", err)
	}

	if lastPosition != nil {
		fmt.Println("last position:", lastPosition)
		initPos := &ubx.MgaIniPos_llh3{
			Lat_dege7: int32(lastPosition.Latitude * 1e7),
			Lon_dege7: int32(lastPosition.Longitude * 1e7),
			PosAcc_cm: 1000 * 100,
		}
		output <- initPos
	}

	timeSet := make(chan time.Time)
	timeGetter := handlers.NewTimeGetter(timeSet)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavPvt, timeGetter)

	done := decoder.Decode(stream)

	now := time.Time{}
	loadAll := false
	select {
	case now = <-timeSet:
		n.handlersRegistry.UnregisterHandler(message.UbxMsgNavPvt, timeGetter)
		sinceStart := time.Since(n.loggerData.GetStartTime())
		err := handlers.SetSystemDate(now)
		if err != nil {
			handleError("setting system date", err)
		}
		newTime := time.Now()
		fmt.Printf("Time set to %s, took %s\n", now, sinceStart)
		startTime := newTime.Add(-sinceStart)
		n.loggerData.SetStartTime(startTime)
		fmt.Println("new start time:", startTime)
	case <-time.After(5 * time.Second):
		fmt.Println("not time yet, will load all ano messages")
		loadAll = true
	}

	if _, err := os.Stat(n.mgaOfflineFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("File %s does not exist\n", n.mgaOfflineFilePath)
	} else {
		go func() {
			loader := handlers.NewAnoLoader()
			n.handlersRegistry.RegisterHandler(message.UbxMsgMgaAckData, loader)
			err = loader.LoadAnoFile(n.mgaOfflineFilePath, loadAll, now, output)
			if err != nil {
				fmt.Println("ERROR loading ano file:", err)
			}
		}()
	}

	fmt.Println("Registering logger ubx message handlers")

	//todo: move all the handlers to the event feed
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavPvt, n.loggerData)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavDop, n.loggerData)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavSat, n.loggerData)
	n.handlersRegistry.RegisterHandler(message.UbxMsgMonRf, n.loggerData)

	if now == (time.Time{}) {
		fmt.Println("Waiting for time")
		now = <-timeSet
		n.handlersRegistry.UnregisterHandler(message.UbxMsgNavPvt, timeGetter)

		fmt.Println("Got time:", now)
		sinceStart := time.Since(n.loggerData.GetStartTime())
		err := handlers.SetSystemDate(now)
		if err != nil {
			handleError("setting system date", err)
		}
		newTime := time.Now()
		fmt.Printf("Time set to %s, took %s\n", now, sinceStart)
		startTime := newTime.Add(-sinceStart)
		n.loggerData.SetStartTime(startTime)
		fmt.Println("new start time:", n.loggerData.GetStartTime())
	}

	n.jsonLogger.StartStoring()
	n.sqliteLogger.StartStoring()

	if err := <-done; err != nil {
		log.Fatalln(err)
	}
}

func handleError(context string, err error) {
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s: %s\n", context, err.Error()))
	}
}
