package main

import (
	"errors"
	"flag"
	"fmt"
	"gnss-logger/logger"
	"gnss-logger/message"
	"gnss-logger/message/handlers"
	"log"
	"os"
	"time"

	"github.com/tarm/serial"
)

func main() {
	databasePath := flag.String("db-path", "/mnt/data/gnss.v1.0.2.db", "path to sqlite database")
	logTTl := flag.Duration("db-log-ttl", 12*time.Hour, "ttl of logs in database")
	jsonDestinationFolder := flag.String("json-destination-folder", "/mnt/data/gps", "json destination folder")
	jsonSaveInterval := flag.Duration("json-save-interval", 15*time.Second, "json save interval")
	jsonDestinationFolderMaxSize := flag.Int64("json-destination-folder-max-size", int64(30000*1024), "json destination folder maximum size") // 30MB
	mgaOfflineFilePath := flag.String("mga-offline-file-path", "/mnt/data/mgaoffline.ubx", "path to mga offline files")
	flag.Parse()

	config := &serial.Config{
		Name:     "/dev/ttyAMA1", //todo: make this configurable???
		Baud:     38400,
		Parity:   serial.ParityNone,
		StopBits: serial.Stop1,
	}

	handlersRegistry := message.NewHandlerRegistry()

	sqlite := logger.NewSqlite(*databasePath)
	err := sqlite.Init(*logTTl)
	handleError("initializing sqlite", err)

	jsonLogger := logger.NewJsonFile(*jsonDestinationFolder, *jsonDestinationFolderMaxSize, *jsonSaveInterval)
	err = jsonLogger.Init()
	handleError("initializing json logger", err)

	loggerData := logger.NewLoggerData(sqlite, jsonLogger)

	decoder := message.NewDecoder(handlersRegistry)

	stream, err := serial.OpenPort(config)
	handleError("opening gps serial port", err)

	timeSet := make(chan time.Time)
	timeGetter := handlers.NewTimeGetter(timeSet)
	handlersRegistry.RegisterHandler(message.UbxMsgNavPvt, timeGetter)

	done := decoder.Decode(stream)

	now := time.Time{}
	loadAll := false
	select {
	case now = <-timeSet:
		handlersRegistry.UnregisterHandler(message.UbxMsgNavPvt, timeGetter)
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
			err = loader.LoadAnoFile(*mgaOfflineFilePath, loadAll, now, stream)
			if err != nil {
				fmt.Println("ERROR loading ano file:", err)
			}
		}()
	}

	fmt.Println("Registering logger ubx message handlers")
	handlersRegistry.RegisterHandler(message.UbxMsgNavPvt, loggerData)
	handlersRegistry.RegisterHandler(message.UbxMsgNavDop, loggerData)
	handlersRegistry.RegisterHandler(message.UbxMsgNavSat, loggerData)

	if now == (time.Time{}) {
		fmt.Println("Waiting for time")
		now = <-timeSet
		fmt.Println("Got time:", now)
		handlersRegistry.UnregisterHandler(message.UbxMsgNavPvt, timeGetter)
	}

	jsonLogger.StartStoring()
	sqlite.StartStoring()

	if err := <-done; err != nil {
		log.Fatalln(err)
	}
}

func handleError(context string, err error) {
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s: %s\n", context, err.Error()))
	}
}
