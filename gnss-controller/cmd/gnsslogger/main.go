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

	startTime := time.Now()

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
	loggerData.SetStartTime(startTime)
	decoder := message.NewDecoder(handlersRegistry)

	stream, err := serial.OpenPort(config)
	handleError("opening gps serial port", err)

	////    def send_cfg_rst(self, reset_type):
	////        """UBX-CFG-RST, reset"""
	////        # Always do a hardware reset
	////        # If on native USB: both Hardware reset (0) and Software reset (1)
	////        # will disconnect and reconnect, giving you a new /dev/tty.
	////        m_data = bytearray(4)
	////        m_data[0] = reset_type & 0xff
	////        m_data[1] = (reset_type >> 8) & 0xff
	////        self.gps_send(6, 0x4, m_data)
	//
	//reset := ubx.CfgRst{
	//	NavBbrMask: 0,
	//	ResetMode:  0,
	//	Reserved1:  0,
	//}

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
