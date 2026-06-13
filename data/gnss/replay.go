package gnss

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Hivemapper/hivemapper-data-logger/logger"
	sensordata "github.com/Hivemapper/hivemapper-data-logger/proto-out"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type GnssReplayDataHandler func(redisKey string, data []byte) error

type GnssReplayFeed struct {
	replayFilePath string
	dataHandler    GnssReplayDataHandler
}

func NewGnssReplayFeed(replayFilePath string, datahandler GnssReplayDataHandler) *GnssReplayFeed {
	g := &GnssReplayFeed{
		replayFilePath,
		datahandler,
	}
	return g
}

// map rediskey to protobuf type
var redisKeyToProto = map[string]proto.Message{
	"NavStatus":  &sensordata.NavStatus{},
	"NavDop":     &sensordata.NavDop{},
	"MonRf":      &sensordata.MonRf{},
	"RxmMeasx":   &sensordata.RxmMeasx{},
	"RxmRawx":    &sensordata.RxmRawx{},
	"RxmSfrbx":   &sensordata.RxmSfrbx{},
	"TimTp":      &sensordata.TimTp{},
	"NavCov":     &sensordata.NavCov{},
	"NavPvt":     &sensordata.NavPvt{},
	"NavPosecef": &sensordata.NavPosecef{},
	"NavVelecef": &sensordata.NavVelecef{},
	"NavSig":     &sensordata.NavSig{},
	"NavTimegps": &sensordata.NavTimegps{},
}

func (f *GnssReplayFeed) Run() error {
	replayFileHandle, err := os.Open(f.replayFilePath)
	if err != nil {
		return fmt.Errorf("opening gnss replay file: %w", err)
	}
	defer replayFileHandle.Close()
	reader := bufio.NewReader(replayFileHandle)

	firstEpoch := true

	// read each line of the file
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("done reading gnss replay file")
				return nil
			}
			return fmt.Errorf("error reading from gnss replay file: %s", err)
		}

		// json with "rediskey" and "data" (pbtxt) entries
		var entry logger.GnssReplayEvent
		err = json.Unmarshal([]byte(line), &entry)
		if err != nil {
			fmt.Printf("error unmarshalling gnss json line: %s\n", err)
			continue
		}
		fmt.Printf("%s\n", entry.RedisKey)

		if entry.RedisKey == "RxmRawx" {
			if !firstEpoch {
				time.Sleep(250 * time.Millisecond)
			}
			firstEpoch = false
		}

		protoType, ok := redisKeyToProto[entry.RedisKey]
		if !ok {
			fmt.Printf("unknown redis key: %s\n", entry.RedisKey)
			continue
		}

		msg := protoType
		err = prototext.Unmarshal([]byte(entry.Data), msg)
		if err != nil {
			fmt.Printf("error unmarshalling gnss data pbtxt: %s\n", err)
			continue
		}

		// now serialize to bytes
		binary_data, err := proto.Marshal(msg)
		if err != nil {
			fmt.Printf("error marshalling gnss data: %s\n", err)
			continue
		}

		f.dataHandler(entry.RedisKey, binary_data)
	}
}
