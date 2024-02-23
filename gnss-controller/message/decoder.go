package message

import (
	"crypto/sha256"

	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/daedaleanai/ublox"
	"github.com/daedaleanai/ublox/nmea"
	"github.com/daedaleanai/ublox/ubx"
	"github.com/streamingfast/shutter"
	"github.com/tarm/serial"

	b64 "encoding/base64"
)

type Decoder struct {
	*shutter.Shutter
	registry *HandlerRegistry
	queue    [][]byte
}

func NewDecoder(registry *HandlerRegistry) *Decoder {
	return &Decoder{
		Shutter:  shutter.New(),
		registry: registry,
	}
}

// Source: NEO-M9N_Integrationmanual_UBX-19015769_C2-Restricted.pdf
//
//	Requires ublox NDA unfortunately.
func validateEcdsaSignature(hash [32]byte, sessionId [24]byte, signature [48]byte) {
	// concatenate hash and session id
	hashAndId := make([]byte, 0)
	hashAndId = append(hashAndId, hash[:]...)
	hashAndId = append(hashAndId, sessionId[:]...)

	// take the sha sum and split it into three parts
	hashAndIdSha256 := sha256.Sum256(hashAndId)
	hashAndIdSha256Part1 := hashAndIdSha256[0:8]
	hashAndIdSha256Part2 := hashAndIdSha256[8:24]
	hashAndIdSha256Part3 := hashAndIdSha256[24:32]

	// xor the first and third part
	hashAndIdSha256Xor := [8]byte{}
	for i := range hashAndIdSha256Xor {
		hashAndIdSha256Xor[i] = hashAndIdSha256Part1[i] ^ hashAndIdSha256Part3[i]
	}

	// append the xor result to the second part of the sha sum
	signatureInput := make([]byte, 0)
	signatureInput = append(signatureInput, hashAndIdSha256Xor[:]...)
	signatureInput = append(signatureInput, hashAndIdSha256Part2...)

	// TODO: validate signatureInput against the gnss-provided signature
}

// compute the hash of the queue contents
func validateHash(sign *ubx.SecEcsign, queue [][]byte) {
	flattened := make([]byte, 0)
	for _, elem := range queue {
		flattened = append(flattened, elem...)
		if elem[0] == 0xB5 {
			fmt.Printf("Entry: %v\n", elem)
		} else {
			fmt.Printf("NMEA Entry: %v\n", elem)
		}
	}
	fmt.Printf("%v messages\n", len(queue))
	fmt.Printf("flattened bytes, length %v\n", len(flattened))

	sum := sha256.Sum256(flattened)
	fmt.Printf("sha256sum: %v\n", sum)

	fmt.Printf("finalhash: %v\n", sign.FinalHash)
	fmt.Printf("messageNum: %v\nsessionId: %v\nversion: %v\n", sign.MsgNum, sign.SessionId, sign.Version)

	fmt.Printf("ecdsa sign: %v", sign.EcdsaSignature)
}

type SecEcsignWithBuffer struct {
	SecEcsign           *ubx.SecEcsign
	Base64MessageBuffer string
}

func encodeBuffer(buffer [][]byte) string {
	flattened := make([]byte, 0)
	for _, message := range buffer {
		// fmt.Printf("message : %v\n", message)
		flattened = append(flattened, message...)
	}
	return b64.StdEncoding.EncodeToString(flattened)
}

type ErrorCallback func(errorMessage string)

func (d *Decoder) Decode(stream *serial.Port, config *serial.Config, errorCallback ErrorCallback) chan error {
	done := make(chan error)
	var ubxDecoder *ublox.Decoder

	initializeDecoder := func() {
		fmt.Println("=========================")
		fmt.Println("Initializing decoder...")
		fmt.Println("=========================")
		// if stream != nil {
		// 	fmt.Println("Closing stream");
		//     stream.Close()
		// }
		stream, err := serial.OpenPort(config)

		if err != nil {
			fmt.Errorf("opening gps serial port: %w", err)
		}
		ubxDecoder = ublox.NewDecoder(stream)
		d.queue = make([][]byte, 0)
	}

	go func() {
		initializeDecoder()

		for {
			if d.IsTerminating() || d.IsTerminated() {
				done <- d.Err()
				break
			}

			//todo: create a cmd to generate a new keypair and store it in the device (for testing purpose). To not loose the public key!
			//Asymmetric signature (private and public keys):
			//need to found if we can get back the public keys (I don't think so)

			//todo: cumulate all bytes between 2 UBX-SEC-ECSIGN messages excluding UBX-SEC-ECSIGN messages(to be validated)
			//todo: get signature from last  UBX-SEC-ECSIGN messages
			//todo: compute hash of all bytes between 2 UBX-SEC-ECSIGN messages
			//todo: refactor ubxDecoder.Decode() func to return all bytes with the message so we can compute the hash
			//todo: signature and computed hash need to be sent with the new data (in the datafeed)
			//todo: add signature and hash to the json log file in the data logger ...
			msg, frame, err := ubxDecoder.Decode()
			if err != nil {
				if err == io.EOF {
					done <- nil
					break
				}
				fmt.Println("WARNING: error decoding ubx", err)
				errorCallback(err.Error())
				initializeDecoder()
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

			if sign, ok := msg.(*ubx.SecEcsign); ok {
				// validateHash(sign, d.queue)

				// hack: swap the original message for this one, which contains the buffer
				secEcsignWithBuffer := SecEcsignWithBuffer{}
				secEcsignWithBuffer.SecEcsign = sign
				secEcsignWithBuffer.Base64MessageBuffer = encodeBuffer(d.queue)
				msg = &secEcsignWithBuffer

				d.queue = make([][]byte, 0)
			} else {
				// add to queue
				if frame[0] == 0xB5 {
					mycopy := make([]byte, len(frame))
					copy(mycopy, frame)
					d.queue = append(d.queue, mycopy)
				} else if string(frame[0]) == "$" {
					mycopy := make([]byte, len(frame))
					copy(mycopy, frame)

					// NMEA Frames are terminated with a "\r\n".
					// For some reason, this is missing in `frame` which was causing
					// hashing not match.
					mycopy = append(mycopy, []byte("\r\n")...)
					d.queue = append(d.queue, mycopy)
				} else {
					fmt.Printf("Unexpected frame type. This might mess with GNSS authentication")
				}
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

func needToRepair(err error) bool {
	// Check if the error message contains the word "unexpected"
	return strings.Contains(err.Error(), "unexpected")
}
