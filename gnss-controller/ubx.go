package gnss_logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/daedaleanai/ublox"
	"github.com/daedaleanai/ublox/ubx"
)

type UbxMessageHandler interface {
	HandleUbxMessage(interface{}) error
}

type AnoLoader struct {
	anoPerSatellite map[uint8]int
	ackChannel      chan *ubx.MgaAckData0
}

func NewAnoLoader() *AnoLoader {
	return &AnoLoader{
		anoPerSatellite: map[uint8]int{},
		ackChannel:      make(chan *ubx.MgaAckData0),
	}
}

func (l *AnoLoader) LoadAnoFile(file string, loadAll bool, now time.Time, stream io.Writer) error {
	fmt.Println("loading mga offline file:", file)
	mgaOfflineFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("opening mga offline file: %w", err)
	}

	mgaOfflineDecoder := ublox.NewDecoder(mgaOfflineFile)
	sentCount := 0
	for {
		msg, err := mgaOfflineDecoder.Decode()
		if err != nil {
			if err == io.EOF {
				fmt.Println("reach mga EOF")
				break
			}
			return fmt.Errorf("decoding mga offline file: %w", err)
		}
		ano := msg.(*ubx.MgaAno)
		anoDate := time.Date(int(ano.Year)+2000, time.Month(ano.Month), int(ano.Day), 0, 0, 0, 0, time.UTC)
		if loadAll || (anoDate.Year() == now.Year() && anoDate.Month() == now.Month() && anoDate.Day() == now.Day()) { //todo: get system date
			fmt.Print(".")
			encoded, err := ubx.Encode(msg.(ubx.Message))
			if err != nil {
				return fmt.Errorf("encoding ano message: %w", err)
			}
			_, err = stream.Write(encoded)
			if err != nil {
				return fmt.Errorf("writing to stream: %w", err)
			}
			time.Sleep(100 * time.Millisecond)

		goAck:
			for {
				select {
				case ack := <-l.ackChannel:
					_, err := json.Marshal(ack)
					if err != nil {
						return err
					}
					break goAck
				case <-time.After(5 * time.Second):
					return errors.New("timeout waiting for ack")
				}
			}
			fmt.Print("!")
			sentCount++
		}
	}

	return nil
}

func (l *AnoLoader) HandleUbxMessage(message interface{}) error {
	ack := message.(*ubx.MgaAckData0)
	l.ackChannel <- ack
	return nil
}
