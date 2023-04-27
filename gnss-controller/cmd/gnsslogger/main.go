package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gnss-logger/gnss"
	"gnss-logger/logger"
	"log"
	"net"
)

func main() {

	sqlite := logger.NewSqlite("/mnt/data/gnss.v1.0.2.db")
	err := sqlite.Init()
	handleError("initializing sqlite", err)

	loggerData := &logger.Data{
		Dop: &logger.Dop{
			GDop: 99.99,
			HDop: 99.99,
			PDop: 99.99,
			TDop: 99.99,
			VDop: 99.99,
			XDop: 99.99,
			YDop: 99.99,
		},
		Satellites: &logger.Satellites{},
	}

	gpsd, err := net.Dial("tcp", "localhost:2947")
	handleError("Dialing gpsd", err)

	_, err = fmt.Fprintf(gpsd, "?WATCH={\"enable\":true,\"json\":true}")
	handleError("Enabling gpsd", err)

	reader := bufio.NewReader(gpsd)
	for {
		buffer, _ := reader.ReadBytes('\n')
		msg := &gnss.Message{}

		err = json.Unmarshal(buffer, &msg)
		handleError("unmarshalling message", err)

		err = msg.UpdateLoggerData(loggerData)
		handleError("updating logger data", err)

		err = sqlite.Log(loggerData)
		handleError("logging data", err)
	}
}

func handleError(context string, err error) {
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s: %s\n", context, err.Error()))
	}
}
