package neom9n

import (
	"errors"
	"fmt"
	"github.com/daedaleanai/ublox/ubx"
	"github.com/streamingfast/gnss-controller/message"
	"github.com/streamingfast/gnss-controller/message/handlers"
	"os"
	"time"

	"github.com/tarm/serial"
)

type Neom9n struct {
	startTime          time.Time
	config             *serial.Config
	handlersRegistry   *message.HandlerRegistry
	decoder            *message.Decoder
	stream             *serial.Port
	output             chan ubx.Message
	mgaOfflineFilePath string
}

func NewNeom9n(serialConfigName string, mgaOfflineFilePath string) *Neom9n {
	n := &Neom9n{
		startTime: time.Now(),
		config: &serial.Config{
			Name:     serialConfigName, // /dev/ttyAMA1
			Baud:     38400,
			Parity:   serial.ParityNone,
			StopBits: serial.Stop1,
		},
		handlersRegistry:   message.NewHandlerRegistry(),
		mgaOfflineFilePath: mgaOfflineFilePath,
		output:             make(chan ubx.Message),
	}

	return n
}

func (n *Neom9n) handleOutputMessages() error {
	for {
		msg := <-n.output
		if _, ok := msg.(*ubx.MonRf); ok {
			encoded, err := ubx.EncodeReq(msg)
			_, err = n.stream.Write(encoded)
			if err != nil {

				return fmt.Errorf("writing message: %w", err)
			}
		}
		encoded, err := ubx.Encode(msg)
		_, err = n.stream.Write(encoded)
		if err != nil {
			return fmt.Errorf("writing message: %w", err)
		}
	}
}

func (n *Neom9n) Init(lastPosition *Position) error {
	n.decoder = message.NewDecoder(n.handlersRegistry)
	stream, err := serial.OpenPort(n.config)
	if err != nil {
		return fmt.Errorf("opening gps serial port: %w", err)
	}

	n.stream = stream
	go func() {
		err = n.handleOutputMessages()
		if err != nil {
			panic(err)
		}
	}()
	
	n.output <- &ubx.CfgValSet{
		Version: 0x00,
		Layers:  ubx.CfgValSetLayers(ubx.CfgValSetLayersRam | ubx.CfgValSetLayersFlash | ubx.CfgValSetLayersBBR),
		CfgData: []*ubx.CfgData{
			{
				Key:   546374490, //0x2091035a CFG-MSGOUT-UBX_MON_RF_UART1
				Value: []byte{0x01},
			},
		},
	}

	n.output <- &ubx.CfgValSet{
		Version: 0x00,
		Layers:  ubx.CfgValSetLayers(ubx.CfgValSetLayersRam | ubx.CfgValSetLayersFlash | ubx.CfgValSetLayersBBR),
		CfgData: []*ubx.CfgData{
			{
				Key:   269549605, //CFG-NAVSPG-ACKAIDING 0x10110025 Acknowledge assistance input messages
				Value: []byte{0x01},
			},
		},
	}

	if lastPosition != nil {
		fmt.Println("last position:", lastPosition)
		initPos := &ubx.MgaIniPos_llh3{
			Lat_dege7: int32(lastPosition.Latitude * 1e7),
			Lon_dege7: int32(lastPosition.Longitude * 1e7),
			PosAcc_cm: 1000 * 100,
		}
		n.output <- initPos
	}

	return nil
}

func (n *Neom9n) Run(dataFeed *DataFeed, timeSetCallback func(now time.Time)) error {
	timeSet := make(chan time.Time)
	timeGetter := handlers.NewTimeGetter(timeSet)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavPvt, timeGetter)

	done := n.decoder.Decode(n.stream)

	now := time.Time{}
	loadAll := false
	select {
	case now = <-timeSet:
		err := n.setSystemStartTime(timeGetter, now)
		if err != nil {
			return fmt.Errorf("setting system start time: %w", err)
		}
		timeSetCallback(n.startTime)
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
			err = loader.LoadAnoFile(n.mgaOfflineFilePath, loadAll, now, n.output)
			if err != nil {
				fmt.Println("ERROR loading ano file:", err)
			}
		}()
	}

	fmt.Println("Registering logger ubx message handlers")
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavPvt, dataFeed)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavDop, dataFeed)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavSat, dataFeed)
	n.handlersRegistry.RegisterHandler(message.UbxMsgMonRf, dataFeed)

	if now == (time.Time{}) {
		fmt.Println("Waiting for time")
		now = <-timeSet
		err := n.setSystemStartTime(timeGetter, now)
		if err != nil {
			return fmt.Errorf("setting system start time: %w", err)
		}

		timeSetCallback(n.startTime)
	}

	if err := <-done; err != nil {
		return err
	}

	return nil
}

func (n *Neom9n) setSystemStartTime(timeGetter *handlers.TimeGetter, now time.Time) error {
	n.handlersRegistry.UnregisterHandler(message.UbxMsgNavPvt, timeGetter)
	sinceStart := time.Since(n.startTime)
	err := handlers.SetSystemDate(now)
	if err != nil {
		return fmt.Errorf("setting system date: %w", err)
	}

	newTime := time.Now()
	fmt.Printf("Time set to %s, took %s\n", now, sinceStart)
	n.startTime = newTime.Add(-sinceStart)
	fmt.Println("new start time:", n.startTime)
	return nil
}

func (n *Neom9n) SetStartTime(startTime time.Time) {
	n.startTime = startTime
}
