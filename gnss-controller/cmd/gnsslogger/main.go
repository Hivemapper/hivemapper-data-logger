package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/streamingfast/gnss-controller/logger"
	"github.com/streamingfast/gnss-controller/message"
	"github.com/streamingfast/gnss-controller/message/handlers"
	"log"
	"os"
	"time"

	"github.com/daedaleanai/ublox/ubx"
	"github.com/tarm/serial"
)

func main() {
	databasePath := flag.String("db-path", "/mnt/data/gnss.v1.0.3.db", "path to sqliteLogger database")
	logTTl := flag.Duration("db-log-ttl", 12*time.Hour, "ttl of logs in database")
	jsonDestinationFolder := flag.String("json-destination-folder", "/mnt/data/gps", "json destination folder")
	jsonSaveInterval := flag.Duration("json-save-interval", 15*time.Second, "json save interval")
	jsonDestinationFolderMaxSize := flag.Int64("json-destination-folder-max-size", int64(30000*1024), "json destination folder maximum size") // 30MB
	mgaOfflineFilePath := flag.String("mga-offline-file-path", "/mnt/data/mgaoffline.ubx", "path to mga offline files")
	flag.Parse()

	startTime := time.Now()

	config := &serial.Config{
		Name:     "/dev/ttyAMA1", //todo: make this configurable???
		Baud:     38400,
		Parity:   serial.ParityNone,
		StopBits: serial.Stop1,
	}

	handlersRegistry := message.NewHandlerRegistry()

	sqliteLogger := logger.NewSqlite(*databasePath)
	err := sqliteLogger.Init(*logTTl)
	handleError("initializing sqliteLogger", err)

	jsonLogger := logger.NewJsonFile(*jsonDestinationFolder, *jsonDestinationFolderMaxSize, *jsonSaveInterval)
	err = jsonLogger.Init()
	handleError("initializing json logger", err)

	loggerData := logger.NewLoggerData(sqliteLogger, jsonLogger)
	loggerData.SetStartTime(startTime)
	decoder := message.NewDecoder(handlersRegistry)

	stream, err := serial.OpenPort(config)
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

	lastPosition, err := sqliteLogger.GetLastPosition()
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
	handlersRegistry.RegisterHandler(message.UbxMsgNavPvt, timeGetter)

	done := decoder.Decode(stream)

	now := time.Time{}
	loadAll := false
	select {
	case now = <-timeSet:
		handlersRegistry.UnregisterHandler(message.UbxMsgNavPvt, timeGetter)
		sinceStart := time.Since(startTime)
		err := handlers.SetSystemDate(now)
		if err != nil {
			handleError("setting system date", err)
		}
		newTime := time.Now()
		fmt.Printf("Time set to %s, took %s\n", now, sinceStart)
		startTime = newTime.Add(-sinceStart)
		loggerData.SetStartTime(startTime)
		fmt.Println("new start time:", startTime)
	case <-time.After(5 * time.Second):
		fmt.Println("not time yet, will load all ano messages")
		loadAll = true
	}

	if _, err := os.Stat(*mgaOfflineFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("File %s does not exist\n", *mgaOfflineFilePath)
	} else {
		go func() {
			loader := handlers.NewAnoLoader()
			handlersRegistry.RegisterHandler(message.UbxMsgMgaAckData, loader)
			err = loader.LoadAnoFile(*mgaOfflineFilePath, loadAll, now, output)
			if err != nil {
				fmt.Println("ERROR loading ano file:", err)
			}
		}()
	}

	fmt.Println("Registering logger ubx message handlers")

	//todo: move all the handlers to the event feed

	handlersRegistry.RegisterHandler(message.UbxMsgNavPvt, loggerData)
	handlersRegistry.RegisterHandler(message.UbxMsgNavDop, loggerData)
	handlersRegistry.RegisterHandler(message.UbxMsgNavSat, loggerData)
	handlersRegistry.RegisterHandler(message.UbxMsgMonRf, loggerData)

	if now == (time.Time{}) {
		fmt.Println("Waiting for time")
		now = <-timeSet
		handlersRegistry.UnregisterHandler(message.UbxMsgNavPvt, timeGetter)

		fmt.Println("Got time:", now)
		sinceStart := time.Since(startTime)
		err := handlers.SetSystemDate(now)
		if err != nil {
			handleError("setting system date", err)
		}
		newTime := time.Now()
		fmt.Printf("Time set to %s, took %s\n", now, sinceStart)
		startTime = newTime.Add(-sinceStart)
		loggerData.SetStartTime(startTime)
		fmt.Println("new start time:", startTime)
	}

	jsonLogger.StartStoring()
	sqliteLogger.StartStoring()

	if err := <-done; err != nil {
		log.Fatalln(err)
	}
}

func handleError(context string, err error) {
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s: %s\n", context, err.Error()))
	}
}