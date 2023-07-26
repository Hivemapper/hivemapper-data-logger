package message

import (
	"encoding/hex"
	"fmt"
	"io"
	"reflect"

	"github.com/daedaleanai/ublox"
	"github.com/daedaleanai/ublox/nmea"
	"github.com/daedaleanai/ublox/ubx"
	"github.com/streamingfast/shutter"
	"github.com/tarm/serial"
)

type Decoder struct {
	*shutter.Shutter
	registry *HandlerRegistry
}

func NewDecoder(registry *HandlerRegistry) *Decoder {
	return &Decoder{
		Shutter:  shutter.New(),
		registry: registry,
	}
}

func (d *Decoder) Decode(stream *serial.Port) chan error {
	done := make(chan error)
	ubxDecoder := ublox.NewDecoder(stream)
	go func() {
		for {
			if d.IsTerminating() || d.IsTerminated() {
				done <- d.Err()
				break
			}
			//todo: cumulate all bytes between 2 UBX-SEC-ECSIGN messages excluding UBX-SEC-ECSIGN messages(to be validated)
			//todo: get signature from last  UBX-SEC-ECSIGN messages
			//todo: compute hash of all bytes between 2 UBX-SEC-ECSIGN messages
			//todo: refactor ubxDecoder.Decode() func to return all bytes with the message so we can compute the hash
			//todo: signature and computed hash need to be sent with the new data (in the datafeed)
			//todo: add signature and hash to the json log file in the data logger ...
			msg, err := ubxDecoder.Decode()
			if err != nil {
				if err == io.EOF {
					done <- nil
					break
				}
				fmt.Println("WARNING: error decoding ubx", err)
				continue
			}
			if txt, ok := msg.(*nmea.TXT); ok {
				fmt.Println("TXT:", txt.Text)
			}
			if cfg, ok := msg.(*ubx.CfgValGet); ok {
				fmt.Println("CFG:", cfg)
			}
			if nack, ok := msg.(*ubx.AckNak); ok {
				fmt.Println("NACK:", nack, hex.EncodeToString([]byte{nack.ClsID, nack.MsgID}))
			}
			d.registry.ForEachHandler(reflect.TypeOf(msg), func(handler UbxMessageHandler) {
				err := handler.HandleUbxMessage(msg)
				if err != nil {
					done <- err
				}
			})
		}
	}()
	return done
}
