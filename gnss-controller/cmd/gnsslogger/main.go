package main

import (
	"flag"
	"fmt"
	"gnss-logger/logger"
	"io"
	"log"
	"time"

	"github.com/daedaleanai/ublox"
	"github.com/tarm/serial"
)

func main() {
	databasePath := flag.String("db-path", "/mnt/data/gnss.v1.0.2.db", "path to sqlite database")
	logTTl := flag.Duration("db-log-ttl", 12*time.Hour, "ttl of logs in database")

	jsonDestinationFolder := flag.String("json-destination-folder", "/mnt/data/gps", "json destination folder")
	jsonSaveInterval := flag.Duration("json-save-interval", 15*time.Second, "json save interval")
	jsonDestinationFolderMaxSize := flag.Int64("json-destination-folder-max-size", int64(30000*1024), "json destination folder maximum size") // 30MB

	flag.Parse()

	sqlite := logger.NewSqlite(*databasePath)
	err := sqlite.Init(*logTTl)
	handleError("initializing sqlite", err)

	jsonLogger := logger.NewJsonFile(*jsonDestinationFolder, *jsonDestinationFolderMaxSize, *jsonSaveInterval)
	err = jsonLogger.Init()
	handleError("initializing json logger", err)

	config := &serial.Config{
		Name:     "/dev/ttyAMA1", //todo: make this configurable???
		Baud:     38400,
		Parity:   serial.ParityNone,
		StopBits: serial.Stop1,
	}

	stream, err := serial.OpenPort(config)
	handleError("opening gps serial port", err)

	d := ublox.NewDecoder(stream)
	loggerData := logger.NewLoggerData()
	for {
		msg, err := d.Decode()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("WARNING: error decoding ubx", err)
			continue
		}
		err = loggerData.HandleMessage(msg)
		handleError("updating logger data", err)

		err = sqlite.Log(loggerData)
		handleError("logging data to sqlite", err)

		err = jsonLogger.Log(loggerData)
		handleError("logging data to json", err)
	}
}

func handleError(context string, err error) {
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s: %s\n", context, err.Error()))
	}
}
