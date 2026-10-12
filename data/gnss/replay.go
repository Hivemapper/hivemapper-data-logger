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

func NewGnssReplayFeed(replayFilePath string, dataHandler GnssReplayDataHandler) *GnssReplayFeed {
	g := &GnssReplayFeed{
		replayFilePath: replayFilePath,
		dataHandler:    dataHandler,
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
	// Pace the replay at the file's own epoch rate rather than a fixed 4 Hz. The interval
	// is learned from consecutive NavPvt itow_ms, so a file generated at any rate replays
	// at that rate. Only the second epoch uses the default, before a delta is known.
	epochSleep := 250 * time.Millisecond
	var prevItowMs uint32

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("done reading gnss replay file")
				return nil
			}
			return fmt.Errorf("error reading from gnss replay file: %s", err)
		}

		var entry logger.GnssReplayEvent
		err = json.Unmarshal([]byte(line), &entry)
		if err != nil {
			fmt.Printf("error unmarshalling gnss json line: %s\n", err)
			continue
		}

		if entry.RedisKey == "RxmRawx" {
			if !firstEpoch {
				time.Sleep(epochSleep)
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

		// Monotonic time and system time in the replayed data needs to be updated
		// so that it will be accepted by bee-sensor-fusion.
		if navPvt, ok := msg.(*sensordata.NavPvt); ok {
			// Learn the epoch interval before the timestamps are rewritten to now.
			if itow := navPvt.GetItowMs(); itow != 0 {
				if prevItowMs != 0 && itow > prevItowMs {
					if d := time.Duration(itow-prevItowMs) * time.Millisecond; d > 0 && d <= 2*time.Second {
						epochSleep = d
					}
				}
				prevItowMs = itow
			}
			navPvt.UptimeMs = logger.MonotonicTime()
			navPvt.SystemTime = time.Now().UTC().Format("2006-01-02 15:04:05.000000")
		}

		binary_data, err := proto.Marshal(msg)
		if err != nil {
			fmt.Printf("error marshalling gnss data: %s\n", err)
			continue
		}

		if err := f.dataHandler(entry.RedisKey, binary_data); err != nil {
			fmt.Printf("error pushing gnss replay data to redis: %s\n", err)
		}
	}
}
