package neom9n

import (
	"time"

	"github.com/Hivemapper/gnss-controller/message"
	"github.com/daedaleanai/ublox/ubx"
)

type Logger interface {
	Log(data *Data) error
}

type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"height"`
}

type Data struct {
	Ttff            int64          `json:"ttff"`
	SystemTime      time.Time      `json:"systemtime"`
	Timestamp       time.Time      `json:"timestamp"`
	Fix             string         `json:"fix"`
	Latitude        float64        `json:"latitude"`
	Longitude       float64        `json:"longitude"`
	Dop             *Dop           `json:"dop"`
	Eph             float64        `json:"eph"` // Estimated horizontal Position (2D) Error in meters. Also known as Estimated Position Error (epe). Certainty unknown.
	TimeResolved    int            `json:"time_resolved"`
	SecEcsign       *ubx.SecEcsign `json:"sec_ecsign"`
	SecEcsignBuffer string         `json:"sec_ecsign_buffer"`
	//todo: add optional signature and hash struct genereated from UBX-SEC-ECSIGN messages by the decoder
}

type Dop struct {
	GDop float64 `json:"gdop"`
	HDop float64 `json:"hdop"`
	PDop float64 `json:"pdop"`
	TDop float64 `json:"tdop"`
	VDop float64 `json:"vdop"`
	XDop float64 `json:"xdop"`
	YDop float64 `json:"ydop"`
}

type DataFeed struct {
	HandleData func(data *Data)
	Data       *Data
}

func NewDataFeed(handleData func(data *Data)) *DataFeed {
	return &DataFeed{
		HandleData: handleData,
		Data:       &Data{},
	}
}

func (df *DataFeed) HandleUbxMessage(msg interface{}) error {
	data := df.Data

	switch m := msg.(type) {
	case *message.SecEcsignWithBuffer:
		data.SecEcsign = m.SecEcsign
		data.SecEcsignBuffer = m.Base64MessageBuffer
		df.HandleData(data)
	}

	return nil
}
