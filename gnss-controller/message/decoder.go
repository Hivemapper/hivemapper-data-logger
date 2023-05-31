package message

import (
	"fmt"
	"io"
	"reflect"

	"github.com/daedaleanai/ublox"
	"github.com/daedaleanai/ublox/nmea"
	"github.com/daedaleanai/ublox/ubx"
	"github.com/tarm/serial"
)

type Decoder struct {
	registry *HandlerRegistry
}

func NewDecoder(registry *HandlerRegistry) *Decoder {
	return &Decoder{
		registry: registry,
	}
}

func (d *Decoder) Decode(stream *serial.Port) chan error {
	done := make(chan error)
	ubxDecoder := ublox.NewDecoder(stream)
	go func() {
		for {
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
			if nack, ok := msg.(*ubx.AckNak); ok {
				fmt.Println("NACK:", nack)
			}

			fmt.Printf("Decoded: %T\n", msg)
			d.registry.ForEachHandler(reflect.TypeOf(msg), func(handler UbxMessageHandler) {
				fmt.Printf("msg value: %v handler: %v\n", reflect.TypeOf(msg), handler)
				err := handler.HandleUbxMessage(msg)
				if err != nil {
					done <- err
				}
			})
		}
	}()
	return done

}
