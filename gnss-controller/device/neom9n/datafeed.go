package neom9n

import (
	"fmt"
	"time"

	"github.com/daedaleanai/ublox/ubx"
)

var noTime = time.Time{}

type Logger interface {
	Log(data *Data) error
}

type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"height"`
}

type Data struct {
	Ttff       int64       `json:"ttff"`
	SystemTime time.Time   `json:"systemtime"`
	Timestamp  time.Time   `json:"timestamp"`
	Fix        string      `json:"fix"`
	Latitude   float64     `json:"latitude"`
	Longitude  float64     `json:"longitude"`
	Altitude   float64     `json:"height"`
	Heading    float64     `json:"heading"`
	Speed      float64     `json:"speed"`
	Dop        *Dop        `json:"dop"`
	Satellites *Satellites `json:"satellites"`
	Sep        float64     `json:"sep"` // Estimated Spherical (3D) Position Error in meters. Guessed to be 95% confidence, but many GNSS receivers do not specify, so certainty unknown.
	Eph        float64     `json:"eph"` // Estimated horizontal Position (2D) Error in meters. Also known as Estimated Position Error (epe). Certainty unknown.
	RF         *RF         `json:"rf,omitempty"`
	startTime  time.Time
}

func (d *Data) GetStartTime() time.Time {
	return d.startTime
}

func (d *Data) SetStartTime(startTime time.Time) {
	d.startTime = startTime
}

func (d *Data) Clone() Data {
	clone := Data{
		Ttff:       d.Ttff,
		SystemTime: d.SystemTime,
		Timestamp:  d.Timestamp,
		Fix:        d.Fix,
		Latitude:   d.Latitude,
		Longitude:  d.Longitude,
		Altitude:   d.Altitude,
		Heading:    d.Heading,
		Speed:      d.Speed,
		Dop: &Dop{
			GDop: d.Dop.GDop,
			HDop: d.Dop.HDop,
			PDop: d.Dop.PDop,
			TDop: d.Dop.TDop,
			VDop: d.Dop.VDop,
			XDop: d.Dop.XDop,
			YDop: d.Dop.YDop,
		},
		Satellites: &Satellites{
			Seen: d.Satellites.Seen,
			Used: d.Satellites.Used,
		},
		Sep: d.Sep,
		Eph: d.Eph,
	}

	if d.RF != nil {
		clone.RF = &RF{
			JammingState: d.RF.JammingState,
			AntStatus:    d.RF.AntStatus,
			AntPower:     d.RF.AntPower,
			PostStatus:   d.RF.PostStatus,
			NoisePerMS:   d.RF.NoisePerMS,
			AgcCnt:       d.RF.AgcCnt,
			JamInd:       d.RF.JamInd,
			OfsI:         d.RF.OfsI,
			MagI:         d.RF.MagI,
			OfsQ:         d.RF.OfsQ,
			MagQ:         d.RF.MagQ,
		}
	}

	return clone
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

type Satellites struct {
	Seen int `json:"seen"`
	Used int `json:"used"`
}

type DataFeed struct {
	handleData func(data *Data)
	Data       *Data
}

func NewDataFeed(handleData func(data *Data)) *DataFeed {
	return &DataFeed{
		handleData: handleData,
		Data: &Data{
			SystemTime: noTime,
			Timestamp:  noTime,
			Dop: &Dop{
				GDop: 99.99,
				HDop: 99.99,
				PDop: 99.99,
				TDop: 99.99,
				VDop: 99.99,
				XDop: 99.99,
				YDop: 99.99,
			},
			Satellites: &Satellites{},
			RF:         &RF{},
		},
	}
}

func (d *DataFeed) GetStartTime() time.Time {
	return d.Data.startTime
}

func (d *DataFeed) SetStartTime(startTime time.Time) {
	d.Data.startTime = startTime
}

// GNSSfix Type: 0: no fix 1: dead reckoning only 2: 2D-fix 3: 3D-fix 4: GNSS + dead reckoning combined 5: time only fix
var fix = []string{"none", "dead reckoning only", "2D", "3D", "GNSS + dead reckoning combined", "time only fix"}

type RF struct {
	JammingState string `json:"jamming_state"`
	AntStatus    string `json:"ant_status"`
	AntPower     string `json:"ant_power"`
	PostStatus   uint32 `json:"post_status"`
	NoisePerMS   uint16 `json:"noise_per_ms"`
	AgcCnt       uint16 `json:"agc_cnt"`
	JamInd       uint8  `json:"jam_ind"`
	OfsI         int8   `json:"ofs_i"`
	MagI         byte   `json:"mag_i"`
	OfsQ         int8   `json:"ofs_q"`
	MagQ         byte   `json:"mag_q"`
}

func (df *DataFeed) HandleUbxMessage(msg interface{}) error {
	data := df.Data
	data.SystemTime = time.Now()

	switch m := msg.(type) {
	case *ubx.NavPvt:
		now := time.Date(int(m.Year_y), time.Month(int(m.Month_month)), int(m.Day_d), int(m.Hour_h), int(m.Min_min), int(m.Sec_s), int(m.Nano_ns), time.UTC)
		data.Timestamp = now
		data.Fix = fix[m.FixType]
		if data.Ttff == 0 && data.Fix == "3D" && data.Dop.HDop < 5.0 {
			fmt.Println("setting ttff to: ", now)
			data.Ttff = time.Since(data.startTime).Milliseconds()
		}
		data.Latitude = float64(m.Lat_dege7) * 1e-7
		data.Longitude = float64(m.Lon_dege7) * 1e-7

		if m.FixType == 3 {
			data.Altitude = float64(m.Height_mm) / 1000 //tv.Althae
		} else {
			data.Altitude = float64(m.HMSL_mm) / 1000 //tv.Althmsl
		}
		data.Eph = float64(m.HAcc_mm) / 1000

		data.Heading = float64(m.HeadMot_dege5) * 1e-5 //tv.HeadMot
		data.Speed = float64(m.GSpeed_mm_s) / 1000     //tv.Speed
	case *ubx.NavDop:
		data.Dop.GDop = float64(m.GDOP) * 0.01
		data.Dop.HDop = float64(m.HDOP) * 0.01
		data.Dop.PDop = float64(m.PDOP) * 0.01
		data.Dop.TDop = float64(m.TDOP) * 0.01
		data.Dop.VDop = float64(m.VDOP) * 0.01
		data.Dop.XDop = float64(m.EDOP) * 0.01
		data.Dop.YDop = float64(m.NDOP) * 0.01

		// we receive NavDop at the end so we handleData here
		df.handleData(data)
	case *ubx.NavSat:
		data.Satellites.Seen = int(m.NumSvs)
		data.Satellites.Used = 0
		for _, sv := range m.Svs {
			if sv.Flags&ubx.NavSatSvUsed != 0x00 {
				data.Satellites.Used++
			}
		}
	case *ubx.MonRf:
		b := m.RFBlocks[0]
		data.RF = &RF{
			JammingState: b.Flags.String(),
			AntStatus:    b.AntStatus.String(),
			AntPower:     b.AntPower.String(),
			PostStatus:   b.PostStatus,
			NoisePerMS:   b.NoisePerMS,
			AgcCnt:       b.AgcCnt,
			JamInd:       b.JamInd,
			OfsI:         b.OfsI,
			MagI:         b.MagI,
			OfsQ:         b.OfsQ,
			MagQ:         b.MagQ,
		}
	}

	return nil
}
