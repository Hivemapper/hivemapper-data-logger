package neom9n

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/daedaleanai/ublox/ubx"
	"github.com/Hivemapper/gnss-controller/message"
	"github.com/Hivemapper/gnss-controller/message/handlers"
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
	decoderDone        chan error
	measxEnabled	   bool
}

func NewNeom9n(serialConfigName string, mgaOfflineFilePath string, initialBaudRate int, measxEnabled bool) *Neom9n {
	n := &Neom9n{
		startTime: time.Now(),
		config: &serial.Config{
			Name: serialConfigName, // /dev/ttyAMA1
			//Baud: 921600,
			Baud:     initialBaudRate,
			Parity:   serial.ParityNone,
			StopBits: serial.Stop1,
		},
		handlersRegistry:   message.NewHandlerRegistry(),
		mgaOfflineFilePath: mgaOfflineFilePath,
		output:             make(chan ubx.Message),
		measxEnabled: measxEnabled,
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

	_ = n.decoder.Decode(n.stream)

	n.delConfig(1079115777, "CFG-UART1-BAUDRATE")
	n.delConfig(807469057, "CFG-RATE-MEAS")
	n.delConfig(269549605, "CFG-NAVSPG-ACKAIDING")
	n.delConfig(546374490, "CFG-MSGOUT-UBX_MON_RF_UART1")
	n.delConfig(546373639, "CFG-MSGOUT-UBX_NAV_PVT_UART1")

	n.setConfig(1079115777, uint32(921600), "CFG-UART1-BAUDRATE") // CFG-UART1-BAUDRATE 0x40520001 The baud rate that should be configured on the UART1

	n.decoder.Shutdown(nil)

	n.config.Baud = 921600
	n.stream.Close()
	n.stream, err = serial.OpenPort(n.config)
	n.decoder = message.NewDecoder(n.handlersRegistry)
	n.decoderDone = n.decoder.Decode(n.stream)

	fmt.Println("Baud changed")

	n.setConfig(269549605, []byte{0x01}, "CFG-NAVSPG-ACKAIDING")         // CFG-NAVSPG-ACKAIDING 0x10110025 Acknowledge assistance input messages
	n.setConfig(546374490, []byte{0x01}, "CFG-MSGOUT-UBX_MON_RF_UART1")  // CFG-MSGOUT-UBX_MON_RF_UART1 0x2091035a Output rate of the UBX-MON-RF message on port UART1
	n.setConfig(546373639, []byte{0x01}, "CFG-MSGOUT-UBX_NAV_PVT_UART1") // CFG-MSGOUT-UBX_NAV_PVT_UART1 0x20910007 Output rate of the UBX-NAV-PVT message on port UART1
	n.setConfig(546373819, []byte{0x01}, "CFG-MSGOUT-NMEA_ID_GGA_UART1")
	if n.measxEnabled {
		n.setConfig(546374149, []byte{0x01}, "CFG-MSGOUT-UBX_RXM_MEASX_UART1")
	}
	//todo: request UBX-SEC-ECSIGN message output
	//todo: find CFG-MSGOUT-UBX_SEC_ECSIGN_UART1

	n.setConfig(807469057, uint16(128), "CFG-RATE-MEAS 0x30210001") // CFG-RATE-MEAS 0x30210001 U2 0.001 s Nominal time between GNSS measurements
	n.setConfig(807469058, uint16(1), "CFG-RATE-NAV")               // CFG-RATE-NAV 0x30210002 Ratio of number of measurements to number of navigation solutions

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

func (n *Neom9n) setConfig(key uint32, value interface{}, description string) {
	n.output <- &ubx.CfgValSet{
		Version: 0x00,
		Layers:  ubx.CfgValSetLayers(ubx.CfgValSetLayersRam | ubx.CfgValSetLayersFlash | ubx.CfgValSetLayersBBR),
		CfgData: []*ubx.CfgData{
			{
				Key:   key,
				Value: value,
			},
		},
	}
	fmt.Println("Set config:", description)
	time.Sleep(100 * time.Millisecond)
}

func (n *Neom9n) getConfig(key uint32) {
	n.output <- &ubx.CfgValGetReq{
		Version: 0x00,
		Layers:  ubx.CfgValSetLayers(ubx.CfgValSetLayersRam),
		Key:     key,
	}
	time.Sleep(100 * time.Millisecond)
}
func (n *Neom9n) delConfig(key uint32, description string) {
	n.output <- &ubx.CfgValDel{
		Layers: ubx.CfgValSetLayers(ubx.CfgValSetLayersRam | ubx.CfgValSetLayersFlash | ubx.CfgValSetLayersBBR),
		Key:    key,
	}

	fmt.Println("Deleted config:", description)
	time.Sleep(100 * time.Millisecond)
}

func (n *Neom9n) Run(dataFeed *DataFeed, timeValidThreshold string, timeSetCallback func(now time.Time)) error {
	timeSet := make(chan time.Time)
	timeGetter := handlers.NewTimeGetter(timeValidThreshold, timeSet)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavPvt, timeGetter)

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
	n.handlersRegistry.RegisterHandler(message.UbxRxmMeasx, dataFeed)
	n.handlersRegistry.RegisterHandler(message.NneaGGA, dataFeed)

	if now == (time.Time{}) {
		fmt.Println("Waiting for time")
		now = <-timeSet
		err := n.setSystemStartTime(timeGetter, now)
		if err != nil {
			return fmt.Errorf("setting system start time: %w", err)
		}

		timeSetCallback(n.startTime)
	}

	if err := <-n.decoderDone; err != nil {
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
