package neom9n

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Hivemapper/gnss-controller/message"
	"github.com/Hivemapper/gnss-controller/message/handlers"
	"github.com/daedaleanai/ublox/ubx"
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
	measxEnabled       bool
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
		measxEnabled:       measxEnabled,
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

	_ = n.decoder.Decode(n.stream, n.config)

	// n.delConfig(1079115777, "CFG-UART1-BAUDRATE")
	// n.delConfig(807469057, "CFG-RATE-MEAS")
	// n.delConfig(269549605, "CFG-NAVSPG-ACKAIDING")
	// n.delConfig(546374490, "CFG-MSGOUT-UBX_MON_RF_UART1")
	// n.delConfig(546373639, "CFG-MSGOUT-UBX_NAV_PVT_UART1")

	n.setConfig(1079115777, uint32(921600), "CFG-UART1-BAUDRATE") // CFG-UART1-BAUDRATE 0x40520001 The baud rate that should be configured on the UART1

	n.decoder.Shutdown(nil)

	n.config.Baud = 921600
	n.stream.Close()
	n.stream, err = serial.OpenPort(n.config)
	n.decoder = message.NewDecoder(n.handlersRegistry)
	n.decoderDone = n.decoder.Decode(n.stream, n.config)

	fmt.Println("===== NEW: Baud changed =====")

	n.setConfig(0x10110025, []byte{0x01}, "CFG-NAVSPG-ACKAIDING") // CFG-NAVSPG-ACKAIDING 0x10110025 Acknowledge assistance input messages

	// set nominal rate of measurements -> navigation solution update rate
	n.setConfig(0x30210001, uint16(100), "CFG-RATE-MEAS 0x30210001") // CFG-RATE-MEAS 0x30210001 U2 0.001 s Nominal time between GNSS measurements
	n.setConfig(0x30210002, uint16(1), "CFG-RATE-NAV")               // CFG-RATE-NAV 0x30210002 Ratio of number of measurements to number of navigation solutions

	// set critical navigation messages to match solution epoch rate
	n.setConfig(0x20910007, []byte{0x01}, "CFG-MSGOUT-UBX_NAV_PVT_UART1") // CFG-MSGOUT-UBX_NAV_PVT_UART1 0x20910007 Output rate of the UBX-NAV-PVT message on port UART1
	n.setConfig(0x2091034b, []byte{0x01}, "CFG-MSGOUT-UBX_SEC_ECSIGN_UART1")
	n.setConfig(0x20910084, []byte{0x01}, "CFG-MSGOUT-UBX_NAV_COV_UART1")
	n.setConfig(0x20910025, []byte{0x01}, "CFG-MSGOUT-UBX_NAV_POSECEF_UART1")
	n.setConfig(0x20910048, []byte{0x01}, "CFG-MSGOUT-UBX_NAV_TIMEGPS_UART1")
	n.setConfig(0x2091003e, []byte{0x01}, "CFG-MSGOUT-UBX_NAV_VELECEF_UART1")
	n.setConfig(0x2091017e, []byte{0x01}, "CFG-MSGOUT-UBX_TIM_TP_UART1")
	n.setConfig(0x2091001b, []byte{0x01}, "CFG-MSGOUT-UBX_NAV_STATUS_UART1")
	n.setConfig(0x20910346, []byte{0x01}, "CFG-MSGOUT-UBX_NAV_SIG_UART1")

	// non critical messages set to 1 Hz
	n.setConfig(0x2091035a, uint8(10), "CFG-MSGOUT-UBX_MON_RF_UART1") // CFG-MSGOUT-UBX_MON_RF_UART1 0x2091035a Output rate of the UBX-MON-RF message on port UART1
	n.setConfig(0x20910635, uint8(10), "CFG-MSGOUT-UBX_SEC_SIG_UART1")
	n.setConfig(0x2091069e, []byte{0x01}, "CFG-MSGOUT-UBX_MON_SYS_UART1")

	// set timepulse configurations
	n.setConfig(0x2005000c, []byte{0x01}, "CFG-TP-TIMEGRID_TP1")
	n.setConfig(0x40050002, uint32(5000), "CFG-TP-PERIOD_TP1")
	n.setConfig(0x40050003, uint32(5000), "CFG-TP-PERIOD_LOCK_TP1")
	n.setConfig(0x40050004, uint32(500), "CFG-TP-LEN_TP1")
	n.setConfig(0x40050005, uint32(500), "CFG-TP-LEN_LOCK_TP1")

	// add raw measurements
	value := []byte{0x00} // default off for MEASX
	if n.measxEnabled {
		value = []byte{0x01}
	}
	n.setConfig(0x20910205, value, "CFG-MSGOUT-UBX_RXM_MEASX_UART1")
	n.setConfig(0x209102a5, uint8(1), "CFG-MSGOUT-UBX_RXM_RAWX_UART1")

	// turn off unneeded messages that are default on
	n.setConfig(0x209100ab, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_RMC_I2C")
	n.setConfig(0x209100af, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_RMC_SPI")
	n.setConfig(0x209100b0, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_VTG_I2C")
	n.setConfig(0x209100b4, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_VTG_SPI")
	n.setConfig(0x209100ba, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_GGA_I2C")
	n.setConfig(0x209100bb, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_GGA_UART1")
	n.setConfig(0x209100be, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_GGA_SPI")
	n.setConfig(0x209100bf, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_GSA_I2C")
	n.setConfig(0x209100c3, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_GSA_SPI")
	n.setConfig(0x209100c4, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_GSV_I2C")
	n.setConfig(0x209100c8, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_GSV_SPI")
	n.setConfig(0x209100c9, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_GLL_I2C")
	n.setConfig(0x209100cd, []byte{0x00}, "CFG-MSGOUT-NMEA_ID_GLL_SPI")
	n.setConfig(0x2091038c, []byte{0x00}, "CFG-MSGOUT-UBX_MON_SPAN_UART1")
	n.setConfig(0x20910061, []byte{0x00}, "CFG-MSGOUT-UBX_NAV_TIMELS_UART1")

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
	fmt.Println("Set config:", description, "value:", value)
	time.Sleep(100 * time.Millisecond)
}

// func (n *Neom9n) getConfig(key uint32) {
// 	n.output <- &ubx.CfgValGetReq{
// 		Version: 0x00,
// 		Layers:  ubx.CfgValSetLayers(ubx.CfgValSetLayersRam),
// 		Key:     key,
// 	}
// 	time.Sleep(100 * time.Millisecond)
// }

// func (n *Neom9n) delConfig(key uint32, description string) {
// 	n.output <- &ubx.CfgValDel{
// 		Layers: ubx.CfgValSetLayers(ubx.CfgValSetLayersRam | ubx.CfgValSetLayersFlash | ubx.CfgValSetLayersBBR),
// 		Key:    key,
// 	}

// 	fmt.Println("Deleted config:", description)
// 	time.Sleep(100 * time.Millisecond)
// }

func (n *Neom9n) Run(dataFeed *DataFeed, redisFeed message.UbxMessageHandler, redisLogsEnabled bool) error {
	now := time.Time{}
	loadAll := true

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
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavCov, dataFeed)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavPosecef, dataFeed)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavTimegps, dataFeed)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavVelecef, dataFeed)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavStatus, dataFeed)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavDop, dataFeed)
	n.handlersRegistry.RegisterHandler(message.UbxMsgNavSat, dataFeed)
	n.handlersRegistry.RegisterHandler(message.UbxMsgMonRf, dataFeed)
	if n.measxEnabled {
		n.handlersRegistry.RegisterHandler(message.UbxRxmMeasx, dataFeed)
	}

	// We need to pass a buffer along with ubx.SecEcsign to the data handler,
	// so we must register a composite class instead of ubx.SecEcsign
	n.handlersRegistry.RegisterHandler(message.UbxSecEcsignWithBuffer, dataFeed)
	n.handlersRegistry.RegisterHandler(message.NneaGGA, dataFeed)

	if redisLogsEnabled {
		fmt.Println("Registering redis handlers")
		n.handlersRegistry.RegisterHandler(message.UbxMsgNavPvt, redisFeed)
		n.handlersRegistry.RegisterHandler(message.UbxMsgNavCov, redisFeed)
		n.handlersRegistry.RegisterHandler(message.UbxMsgNavPosecef, redisFeed)
		n.handlersRegistry.RegisterHandler(message.UbxMsgNavTimegps, redisFeed)
		n.handlersRegistry.RegisterHandler(message.UbxMsgNavVelecef, redisFeed)
		n.handlersRegistry.RegisterHandler(message.UbxMsgNavStatus, redisFeed)
		n.handlersRegistry.RegisterHandler(message.UbxMsgNavDop, redisFeed)
		n.handlersRegistry.RegisterHandler(message.UbxMsgNavSat, redisFeed)
		n.handlersRegistry.RegisterHandler(message.UbxMsgMonRf, redisFeed)
		if n.measxEnabled {
			n.handlersRegistry.RegisterHandler(message.UbxRxmMeasx, redisFeed)
		}
	} else {
		fmt.Println("Redis handler not set")
	}

	if err := <-n.decoderDone; err != nil {
		return err
	}

	return nil
}
