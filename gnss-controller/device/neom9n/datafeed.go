package neom9n

import (
	"fmt"
	"math"
	"time"

	"github.com/Hivemapper/gnss-controller/message"
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
	Ttff                int64       `json:"ttff"`
	SystemTime          time.Time   `json:"systemtime"`
	ActualSystemTime    time.Time   `json:"actual_systemtime"`
	Timestamp           time.Time   `json:"timestamp"`
	Fix                 string      `json:"fix"`
	Latitude            float64     `json:"latitude"`
	UnfilteredLatitude  float64     `json:"unfiltered_latitude"`
	Longitude           float64     `json:"longitude"`
	UnfilteredLongitude float64     `json:"unfiltered_longitude"`
	Altitude            float64     `json:"height"`
	Heading             float64     `json:"heading"`
	Speed               float64     `json:"speed"`
	Dop                 *Dop        `json:"dop"`
	Satellites          *Satellites `json:"satellites"`
	Sep                 float64     `json:"sep"` // Estimated Spherical (3D) Position Error in meters. Guessed to be 95% confidence, but many GNSS receivers do not specify, so certainty unknown.
	Eph                 float64     `json:"eph"` // Estimated horizontal Position (2D) Error in meters. Also known as Estimated Position Error (epe). Certainty unknown.
	Cno                 float64     `json:"cno"`
	PosConfidence       float64     `json:"pos_confidence"`
	RF                  *RF         `json:"rf,omitempty"`
	SpeedAccuracy       float64     `json:"speed_accuracy"`
	HeadingAccuracy     float64     `json:"heading_accuracy"`
	TimeResolved        int         `json:"time_resolved"`
	HorizontalAccuracy  float64     `json:"horizontal_accuracy"`
	VerticalAccuracy    float64     `json:"vertical_accuracy"`

	startTime       time.Time
	GGA             string         `json:"gga"`
	RxmMeasx        *ubx.RxmMeasx  `json:"rxm_measx"`
	SecEcsign       *ubx.SecEcsign `json:"sec_ecsign"`
	SecEcsignBuffer string         `json:"sec_ecsign_buffer"`
	//todo: add optional signature and hash struct genereated from UBX-SEC-ECSIGN messages by the decoder
}

func (d *Data) Clone() Data {
	clone := Data{
		Ttff:                d.Ttff,
		SystemTime:          d.SystemTime,
		ActualSystemTime:    d.ActualSystemTime,
		Timestamp:           d.Timestamp,
		Fix:                 d.Fix,
		Latitude:            d.Latitude,
		UnfilteredLatitude:  d.UnfilteredLatitude,
		Longitude:           d.Longitude,
		UnfilteredLongitude: d.UnfilteredLongitude,
		Altitude:            d.Altitude,
		Heading:             d.Heading,
		Speed:               d.Speed,
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
		Sep:                d.Sep,
		Eph:                d.Eph,
		TimeResolved:       d.TimeResolved,
		SpeedAccuracy:      d.SpeedAccuracy,
		HeadingAccuracy:    d.HeadingAccuracy,
		HorizontalAccuracy: d.HorizontalAccuracy,
		VerticalAccuracy:   d.VerticalAccuracy,
		GGA:                d.GGA,
		Cno:                d.Cno,
		PosConfidence:      d.PosConfidence,
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
	if d.RxmMeasx != nil {
		clone.RxmMeasx = &ubx.RxmMeasx{
			Reserved1:       d.RxmMeasx.Reserved1,
			GpsTOW_ms:       d.RxmMeasx.GpsTOW_ms,
			GloTOW_ms:       d.RxmMeasx.GloTOW_ms,
			BdsTOW_ms:       d.RxmMeasx.BdsTOW_ms,
			Reserved2:       d.RxmMeasx.Reserved2,
			QzssTOW_ms:      d.RxmMeasx.QzssTOW_ms,
			GpsTOWacc_msl4:  d.RxmMeasx.GpsTOWacc_msl4,
			GloTOWacc_msl4:  d.RxmMeasx.GloTOWacc_msl4,
			BdsTOWacc_msl4:  d.RxmMeasx.BdsTOWacc_msl4,
			Reserved3:       d.RxmMeasx.Reserved3,
			QzssTOWacc_msl4: d.RxmMeasx.QzssTOWacc_msl4,
			NumSV:           d.RxmMeasx.NumSV,
			Flags:           d.RxmMeasx.Flags,
			Reserved4:       d.RxmMeasx.Reserved4,
		}

		for _, sv := range d.RxmMeasx.SV {
			csv := &ubx.RxmMeasxSVType{
				GnssId:          sv.GnssId,
				SvId:            sv.SvId,
				CNo:             sv.CNo,
				MpathIndic:      sv.MpathIndic,
				DopplerMS_m_s:   sv.DopplerMS_m_s,
				DopplerHz_hz:    sv.DopplerHz_hz,
				WholeChips:      sv.WholeChips,
				FracChips:       sv.FracChips,
				CodePhase_msl21: sv.CodePhase_msl21,
				IntCodePhase_ms: sv.IntCodePhase_ms,
				PseuRangeRMSErr: sv.PseuRangeRMSErr,
				Reserved5:       sv.Reserved5,
			}
			clone.RxmMeasx.SV = append(clone.RxmMeasx.SV, csv)
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
	HandleData func(data *Data)
	Data       *Data
}

func NewDataFeed(handleData func(data *Data)) *DataFeed {
	return &DataFeed{
		HandleData: handleData,
		Data: &Data{
			SystemTime:       noTime,
			ActualSystemTime: noTime,
			Timestamp:        noTime,
			Dop: &Dop{
				GDop: 99.99,
				HDop: 99.99,
				PDop: 99.99,
				TDop: 99.99,
				VDop: 99.99,
				XDop: 99.99,
				YDop: 99.99,
			},
			Satellites:   &Satellites{},
			TimeResolved: 0,
			RF:           &RF{},
		},
	}
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

const (
	diffThreshold = 200
	fixThreshold  = 40
)

var (
	prevTime       time.Time
	prevSystemTime time.Time
	fixedTimes     int64
	prevItowMs     uint32
)

func (df *DataFeed) HandleUbxMessage(msg interface{}) error {
	data := df.Data

	switch m := msg.(type) {
	case *ubx.NavPvt:
		now := time.Date(int(m.Year_y), time.Month(int(m.Month_month)), int(m.Day_d), int(m.Hour_h), int(m.Min_min), int(m.Sec_s), int(m.Nano_ns), time.UTC)
		data.Timestamp = now

		data.ActualSystemTime = time.Now().UTC()
		data.SystemTime = time.Now().UTC()

		if m.Valid == ubx.NavPvtFullyResolved {
			data.TimeResolved = 0
		} else {
			data.TimeResolved = 1
		}

		var timeDiff, systemTimeDiff int64
		if !prevTime.IsZero() {
			timeDiff = data.Timestamp.Sub(prevTime).Milliseconds()
		}
		if !prevSystemTime.IsZero() {
			systemTimeDiff = data.SystemTime.Sub(prevSystemTime).Milliseconds()
		}

		if timeDiff < diffThreshold && systemTimeDiff > diffThreshold {
			data.SystemTime = prevSystemTime.Add(time.Duration(timeDiff) * time.Millisecond)
			fixedTimes++
			if fixedTimes > fixThreshold {
				data.SystemTime = time.Now().UTC()
				fixedTimes = 0
			}
		} else {
			fixedTimes = 0
		}

		prevTime = data.Timestamp
		prevSystemTime = data.SystemTime

		data.Fix = fix[m.FixType]
		if data.Ttff == 0 && data.Fix == "3D" && data.Dop.HDop < 5.0 {
			fmt.Println("setting ttff to: ", now)
			data.Ttff = time.Since(data.startTime).Milliseconds()
		}
		data.Latitude = float64(m.Lat_dege7) * 1e-7
		data.Longitude = float64(m.Lon_dege7) * 1e-7
		data.UnfilteredLatitude = float64(m.Lat_dege7) * 1e-7
		data.UnfilteredLongitude = float64(m.Lon_dege7) * 1e-7

		if m.FixType == 3 {
			data.Altitude = float64(m.Height_mm) / 1000 //tv.Althae
		} else {
			data.Altitude = float64(m.HMSL_mm) / 1000 //tv.Althmsl
		}
		data.Eph = float64(m.HAcc_mm) / 1000

		data.Heading = float64(m.HeadMot_dege5) * 1e-5 //tv.HeadMot
		data.HeadingAccuracy = float64(m.HeadAcc_dege5) * 1e-5

		data.Speed = float64(m.GSpeed_mm_s) / 1000 //tv.Speed
		data.SpeedAccuracy = float64(m.SAcc_mm_s) / 1000

		data.HorizontalAccuracy = float64(m.HAcc_mm) / 1000
		data.VerticalAccuracy = float64(m.VAcc_mm) / 1000

		if prevItowMs != 0 && m.ITOW_ms-prevItowMs > 125 {
			fmt.Println("[WARNING] NavPvt drop of", m.ITOW_ms-prevItowMs, "ms (", prevItowMs, ",", m.ITOW_ms, ")")
		}
		prevItowMs = m.ITOW_ms

	case *ubx.NavDop:
		data.Dop.GDop = float64(m.GDOP) * 0.01
		data.Dop.HDop = float64(m.HDOP) * 0.01
		data.Dop.PDop = float64(m.PDOP) * 0.01
		data.Dop.TDop = float64(m.TDOP) * 0.01
		data.Dop.VDop = float64(m.VDOP) * 0.01
		data.Dop.XDop = float64(m.EDOP) * 0.01
		data.Dop.YDop = float64(m.NDOP) * 0.01

	case *ubx.NavSat:
		data.Satellites.Seen = int(m.NumSvs)
		data.Satellites.Used = 0
		data.Cno = 0
		data.PosConfidence = 0
		cno := 0.0
		max_residual := 0.0
		for _, sv := range m.Svs {
			if sv.Flags&ubx.NavSatSvUsed != 0x00 {
				data.Satellites.Used++
				cno += float64(sv.Cno_dbhz)
				if math.Abs(float64(sv.PrRes_me1)*0.1) > max_residual {
					max_residual = math.Abs(float64(sv.PrRes_me1) * 0.1)
				}
			}
		}
		if data.Satellites.Used > 0 {
			data.Cno = float64(cno) / float64(data.Satellites.Used)

			// compute confidence as y = 1.05^-|x| where x is the maximum residual [m]
			data.PosConfidence = math.Pow(1.05, -max_residual)
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
	case *ubx.RxmMeasx:
		data.RxmMeasx = m
	case *message.SecEcsignWithBuffer:
		data.SecEcsign = m.SecEcsign
		data.SecEcsignBuffer = m.Base64MessageBuffer
		df.HandleData(data)
	case *ubx.NavEoe:
		// we receive NavEoe message at the end of epoch so handle data here
		clone := data.Clone()
		df.HandleData(&clone)
	}

	return nil
}
