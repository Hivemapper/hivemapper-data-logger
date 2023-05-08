package main

import (
	"errors"
	"flag"
	"fmt"
	gnss_logger "gnss-logger"
	"gnss-logger/logger"
	"io"
	"log"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/daedaleanai/ublox"
	"github.com/daedaleanai/ublox/ubx"
	"github.com/tarm/serial"
)

func main() {
	databasePath := flag.String("db-path", "/mnt/data/gnss.v1.0.2.db", "path to sqlite database")
	logTTl := flag.Duration("db-log-ttl", 12*time.Hour, "ttl of logs in database")

	jsonDestinationFolder := flag.String("json-destination-folder", "/mnt/data/gps", "json destination folder")
	jsonSaveInterval := flag.Duration("json-save-interval", 15*time.Second, "json save interval")
	jsonDestinationFolderMaxSize := flag.Int64("json-destination-folder-max-size", int64(30000*1024), "json destination folder maximum size") // 30MB

	mgaOfflineFilePath := flag.String("mga-offline-file-path", "/mnt/data/mgaoffline.ubx", "path to mga offline files")

	timeSet := make(chan time.Time)
	timeGetter := gnss_logger.NewTimeGetter(timeSet)

	messageHandlersLock := sync.Mutex{}
	messageHandlers := map[reflect.Type][]gnss_logger.UbxMessageHandler{}
	messageHandlers[reflect.TypeOf(&ubx.NavPvt{})] = []gnss_logger.UbxMessageHandler{timeGetter}

	flag.Parse()

	sqlite := logger.NewSqlite(*databasePath)
	err := sqlite.Init(*logTTl)
	handleError("initializing sqlite", err)

	jsonLogger := logger.NewJsonFile(*jsonDestinationFolder, *jsonDestinationFolderMaxSize, *jsonSaveInterval)
	err = jsonLogger.Init()
	handleError("initializing json logger", err)

	loggerData := logger.NewLoggerData(sqlite, jsonLogger)

	config := &serial.Config{
		Name:     "/dev/ttyAMA1", //todo: make this configurable???
		Baud:     38400,
		Parity:   serial.ParityNone,
		StopBits: serial.Stop1,
	}

	stream, err := serial.OpenPort(config)
	handleError("opening gps serial port", err)

	done := make(chan error)
	go func() {
		d := ublox.NewDecoder(stream)
		for {
			msg, err := d.Decode()
			if err != nil {
				if err == io.EOF {
					done <- nil
					break
				}
				fmt.Println("WARNING: error decoding ubx", err)
				continue
			}

			messageHandlersLock.Lock()
			handlers := messageHandlers[reflect.TypeOf(msg)]
			for _, handler := range handlers {
				err := handler.HandleUbxMessage(msg)
				if err != nil {
					done <- err
				}
			}
			messageHandlersLock.Unlock()
		}
	}()

	now := time.Time{}
	loadAll := false
	select {
	case now = <-timeSet:
		messageHandlersLock.Lock()
		var newHandlers []gnss_logger.UbxMessageHandler
		handlers := messageHandlers[reflect.TypeOf(&ubx.NavPvt{})]
		for _, handler := range handlers {
			if _, ok := handler.(*gnss_logger.TimeGetter); ok {
				fmt.Println("unregistering time getter")
				continue
			}
			newHandlers = append(newHandlers, handler)
		}
		messageHandlers[reflect.TypeOf(&ubx.NavPvt{})] = newHandlers
		messageHandlersLock.Unlock()

	case <-time.After(5 * time.Second):
		fmt.Println("not time yet, will load all ano messages")
		loadAll = true
	}

	if _, err := os.Stat(*mgaOfflineFilePath); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("File %s does not exist\n", *mgaOfflineFilePath)
	} else {
		go func() {
			loader := gnss_logger.NewAnoLoader()
			messageHandlers[reflect.TypeOf(&ubx.MgaAckData0{})] = []gnss_logger.UbxMessageHandler{loader}
			err = loader.LoadAnoFile(*mgaOfflineFilePath, loadAll, now, stream)
			if err != nil {
				fmt.Println("ERROR loading ano file:", err)
			}
		}()
	}

	if now == (time.Time{}) {
		fmt.Println("Waiting for time")
		now = <-timeSet
		messageHandlersLock.Lock()
		var newHandlers []gnss_logger.UbxMessageHandler
		handlers := messageHandlers[reflect.TypeOf(&ubx.NavPvt{})]
		for _, handler := range handlers {
			if _, ok := handler.(*gnss_logger.TimeGetter); ok {
				fmt.Println("unregistering time getter")
				continue
			}
			newHandlers = append(newHandlers, handler)
		}
		messageHandlers[reflect.TypeOf(&ubx.NavPvt{})] = newHandlers
		messageHandlersLock.Unlock()

		fmt.Println("Got time:", now)
	}
	fmt.Println("Registering logger ubx message handlers")
	messageHandlers[reflect.TypeOf(&ubx.NavPvt{})] = []gnss_logger.UbxMessageHandler{loggerData}
	messageHandlers[reflect.TypeOf(&ubx.NavDop{})] = []gnss_logger.UbxMessageHandler{loggerData}
	messageHandlers[reflect.TypeOf(&ubx.NavSat{})] = []gnss_logger.UbxMessageHandler{loggerData}

	if err := <-done; err != nil {
		log.Fatalln(err)
	}
}

func handleError(context string, err error) {
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s: %s\n", context, err.Error()))
	}
}
