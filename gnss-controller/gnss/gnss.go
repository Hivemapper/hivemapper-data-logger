package gnss

import (
	"encoding/json"
	"fmt"
	"gnss-logger/logger"
	"time"
)

var fix = []string{"unkown", "no fix", "2D", "3D"}
var satSystem = []string{"GPS", "unknown", "Galileo", "Beidou", "unknown", "QZSS", "GLONASS"}

type FixMode byte

const (
	UnknownFix FixMode = iota
	NoFix
	Fix2D
	Fix3D
)

type LoggerDataUpdater interface {
	UpdateLoggerData(*logger.Data) error
}

type Message struct {
	Class  *ReportClass `json:"class"`
	Report LoggerDataUpdater
}

type ReportClass struct {
	Class string `json:"class"`
}

type SatInfo struct {
	PRNNumber        int     `json:"prn"`              // SV PRN number, pseudo-random noise or gold code
	Elevation        float64 `json:"elevation"`        // Elevation in degrees, 90 maximum
	Azimuth          float64 `json:"azimuth"`          // Azimuth, degrees from true north, 000 to 359
	SignalNoiseRatio float64 `json:"signalNoiseRatio"` // SNR, 00-99 dB (null when not tracking)
	Used             bool    `json:"used"`             // Used in current solution? (SBAS/WAAS/EGNOS satellites may be flagged used if the solution has corrections from them, but not all drivers make this information available.)
	Type             string  `json:"type"`             // Satellite System
	Svid             float64 `json:"svid"`             // The satellite ID within its constellation. As defined by u-blox, not NMEA).
	Sigid            float64 `json:"sigid"`            // The signal ID of this signal. As defined by u-blox, not NMEA. See u-blox doc for details.
	Freqid           int     `json:"freqid"`           // For GLONASS satellites only: the frequency ID of the signal. As defined by u-blox, range 0 to 13. The freqid is the frequency slot plus 7.
	Health           int     `json:"health"`           // The health of this satellite. 0 is unknown, 1 is OK, and 2 is unhealthy.
}

func (si *SatInfo) UpdateLoggerData(loggerData *logger.Data) error {
	return nil
}

// Satellite describes a location of a GPS satellite
type Satellite struct {
	PRN    int     `json:"PRN"`    // PRN ID of the satellite. 1-63 are GNSS satellites, 64-96 are GLONASS satellites, 100-164 are SBAS satellites
	Az     float64 `json:"az"`     // Azimuth, degrees from true north.
	El     float64 `json:"el"`     // Elevation in degrees.
	Ss     float64 `json:"ss"`     // Signal to Noise ratio in dBHz.
	Used   bool    `json:"used"`   // Used in current solution? (SBAS/WAAS/EGNOS satellites may be flagged used if the solution has corrections from them, but not all drivers make this information available.)
	Gnssid uint8   `json:"gnssid"` // The GNSS ID, as defined by u-blox, not NMEA. 0=GPS, 2=Galileo, 3=Beidou, 5=QZSS, 6-GLONASS.
	Svid   float64 `json:"svid"`   // The satellite ID within its constellation. As defined by u-blox, not NMEA).
	Sigid  float64 `json:"sigid"`  // The signal ID of this signal. As defined by u-blox, not NMEA. See u-blox doc for details.
	Freqid int     `json:"freqid"` // For GLONASS satellites only: the frequency ID of the signal. As defined by u-blox, range 0 to 13. The freqid is the frequency slot plus 7.
	Health int     `json:"health"` // The health of this satellite. 0 is unknown, 1 is OK, and 2 is unhealthy.
}

func (s *Satellite) UpdateLoggerData(loggerData *logger.Data) error {
	return nil
}

// TPVReport is a Time-Position-Velocity report
type TPVReport struct {
	Device      string    `json:"device"`      // Name of the originating device.
	Mode        FixMode   `json:"mode"`        // NMEA mode: NMEA mode: 0=unknown, 1=no fix, 2=2D, 3=3D.
	Status      float64   `json:"status"`      // GPS fix status: 0=Unknown, 1=Normal, 2=DGPS, 3=RTK Fixed, 4=RTK Floating, 5=DR, 6=GNSSDR, 7=Time (surveyed), 8=Simulated, 9=P(Y)
	Time        time.Time `json:"time"`        // Time/date stamp in ISO8601 format, UTC. May have a fractional part of up to .001sec precision. May be absent if the mode is not 2D or 3D. May be present, but invalid, if there is no fix. Verify 3 consecutive 3D fixes before believing it is UTC. Even then it may be off by several seconds until the current leap seconds is known.
	Althae      float64   `json:"altHAE"`      // Altitude, height above ellipsoid, in meters. Probably WGS84.
	Altmsl      float64   `json:"altMSL"`      // MSL Altitude in meters. The geoid used is rarely specified and is often inaccurate. See the comments below on geoidSep. altMSL is altHAE minus geoidSep.
	Alt         float64   `json:"alt"`         // Deprecated. Undefined. Use altHAE or altMSL.
	Climb       float64   `json:"climb"`       // Climb (positive) or sink (negative) rate, meters per second.
	Datum       string    `json:"datum"`       // Current datum. Hopefully WGS84.
	Depth       float64   `json:"depth"`       // Depth in meters. Probably depth below the keel…​
	Dgpsage     float64   `json:"dgpsAge"`     // Age of DGPS data. In seconds
	Dgpssta     float64   `json:"dgpsSta"`     // Station of DGPS data.
	Epc         float64   `json:"epc"`         // Estimated climb error in meters per second. Certainty unknown.
	Epd         float64   `json:"epd"`         // Estimated track (direction) error in degrees. Certainty unknown.
	Eph         float64   `json:"eph"`         // Estimated horizontal Position (2D) Error in meters. Also known as Estimated Position Error (epe). Certainty unknown.
	Eps         float64   `json:"eps"`         // Estimated speed error in meters per second. Certainty unknown.
	Ept         float64   `json:"ept"`         // Estimated time stamp error in seconds. Certainty unknown.
	Epx         float64   `json:"epx"`         // Longitude error estimate in meters. Certainty unknown.
	Epy         float64   `json:"epy"`         // Latitude error estimate in meters. Certainty unknown.
	Epv         float64   `json:"epv"`         // Estimated vertical error in meters. Certainty unknown.
	Geoidsep    float64   `json:"geoidSep"`    // Geoid separation is the difference between the WGS84 reference ellipsoid and the geoid (Mean Sea Level) in meters. Almost no GNSS receiver specifies how they compute their geoid. gpsd interpolates the geoid from a 5x5 degree table of EGM2008 values when the receiver does not supply a geoid separation. The gpsd computed geoidSep is usually within one meter of the "true" value, but can be off as much as 12 meters.
	Lat         float64   `json:"lat"`         // Latitude in degrees: +/- signifies North/South.
	Leapseconds int       `json:"leapseconds"` // Current leap seconds.
	Lon         float64   `json:"lon"`         // Longitude in degrees: +/- signifies East/West.
	Track       float64   `json:"track"`       // Course over ground, degrees from true north.
	Magtrack    float64   `json:"magtrack"`    // Course over ground, degrees magnetic.
	Magvar      float64   `json:"magvar"`      // Magnetic variation, degrees. Also known as the magnetic declination (the direction of the horizontal component of the magnetic field measured clockwise from north) in degrees, Positive is West variation. Negative is East variation.
	Speed       float64   `json:"speed"`       // Speed over ground, meters per second.
	Ecefx       float64   `json:"ecefx"`       // ECEF X position in meters.
	Ecefy       float64   `json:"ecefy"`       // ECEF Y position in meters.
	Ecefz       float64   `json:"ecefz"`       // ECEF Z position in meters.
	Ecefpacc    float64   `json:"ecefpAcc"`    // ECEF position error in meters. Certainty unknown.
	Ecefvx      float64   `json:"ecefvx"`      // ECEF X velocity in meters per second.
	Ecefvy      float64   `json:"ecefvy"`      // ECEF Y velocity in meters per second.
	Ecefvz      float64   `json:"ecefvz"`      // ECEF Z velocity in meters per second.
	Ecefvacc    float64   `json:"ecefvAcc"`    // ECEF velocity error in meters per second. Certainty unknown.
	Sep         float64   `json:"sep"`         // Estimated Spherical (3D) Position Error in meters. Guessed to be 95% confidence, but many GNSS receivers do not specify, so certainty unknown.
	Reld        float64   `json:"relD"`        // Down component of relative position vector in meters.
	Rele        float64   `json:"relE"`        // East component of relative position vector in meters.
	Reln        float64   `json:"relN"`        // North component of relative position vector in meters.
	Veld        float64   `json:"velD"`        // Down velocity component in meters.
	Vele        float64   `json:"velE"`        // East velocity component in meters.
	Veln        float64   `json:"velN"`        // North velocity component in meters.
	Wanglem     float64   `json:"wanglem"`     // Wind angle magnetic in degrees.
	Wangler     float64   `json:"wangler"`     // Wind angle relative in degrees.
	Wanglet     float64   `json:"wanglet"`     // Wind angle true in degrees.
	Wspeedr     float64   `json:"wspeedr"`     // Wind speed relative in meters per second.
	Wspeedt     float64   `json:"wspeedt"`     // Wind speed true in meters per second.
}

func (tv *TPVReport) UpdateLoggerData(d *logger.Data) error {
	d.Fix = fix[tv.Mode]
	d.Latitude = tv.Lat
	d.Longitude = tv.Lon

	if tv.Mode == Fix3D {
		d.Altitude = tv.Althae
	} else {
		d.Altitude = tv.Altmsl
	}
	d.Sep = tv.Sep
	d.Eph = tv.Eph
	d.Heading = tv.Track
	d.Speed = tv.Speed
	return nil
}

// SKYReport reports sky view of GPS satellites
type SKYReport struct {
	Tag        string       `json:"tag"`
	Device     string       `json:"device"`
	Time       time.Time    `json:"time"`
	Xdop       float64      `json:"xdop"`
	Ydop       float64      `json:"ydop"`
	Vdop       float64      `json:"vdop"`
	Tdop       float64      `json:"tdop"`
	Hdop       float64      `json:"hdop"`
	Pdop       float64      `json:"pdop"`
	Gdop       float64      `json:"gdop"`
	Satellites []*Satellite `json:"satellites"`
}

func (sr *SKYReport) UpdateLoggerData(d *logger.Data) error {
	dop := &logger.Dop{
		GDop: sr.Gdop,
		HDop: sr.Hdop,
		PDop: sr.Pdop,
		TDop: sr.Tdop,
		VDop: sr.Vdop,
		XDop: sr.Xdop,
		YDop: sr.Ydop,
	}
	d.Dop = dop
	d.Timestamp = sr.Time

	sats := &logger.Satellites{}
	for _, sat := range sr.Satellites {
		if sat.Used {
			sats.Used++
		}
		sats.Seen++
	}

	d.Satellites = sats

	return nil
}

type PPSReport struct {
	Device string `json:"device"` // Name of the originating device
	TimeData
}

func (sr *PPSReport) UpdateLoggerData(loggerData *logger.Data) error {
	return nil
}

type TimeData struct {
	RealSec     float64   `json:"real_sec"`    // seconds from the PPS source
	RealNsec    float64   `json:"real_nsec"`   // nanoseconds from the PPS source
	ClockSec    float64   `json:"clock_sec"`   // seconds from the system clock
	ClockNsec   float64   `json:"clock_nsec"`  // nanoseconds from the system clock
	Precision   float64   `json:"precision"`   // NTP style estimate of PPS precision
	Shm         string    `json:"shm"`         // shm key of this PPS
	Qerr        float64   `json:"qErr"`        // Quantization error of the PPS, in picoseconds. Sometimes called the "sawtooth" error.
	Timestamp   time.Time `json:"timestamp"`   // Time Object
	Leapseconds int       `json:"leapseconds"` // leapseconds between GPS-Time and UTC
}

func (b *Message) UnmarshalJSON(data []byte) error {
	class := &ReportClass{}
	json.Unmarshal(data, &class)
	if err := json.Unmarshal(data, &class); err != nil {
		return err
	}
	b.Class = class

	switch class.Class {
	case "TPV":
		b.Report = &TPVReport{}
	case "SKY":
		b.Report = &SKYReport{}
	case "PPS":
		b.Report = &PPSReport{}
	case "VERSION":
	case "DEVICES":
	case "WATCH":
	default:
		return fmt.Errorf("unknown message class: %s", class.Class)
	}

	if b.Report == nil {
		return nil
	}
	err := json.Unmarshal(data, &b.Report)
	if err != nil {
		return fmt.Errorf("unmarshalling %s: %w", class.Class, err)
	}

	return nil
}

func (b *Message) UpdateLoggerData(data *logger.Data) error {
	if b.Report == nil {
		return nil
	}
	data.SystemTime = time.Now()
	return b.Report.UpdateLoggerData(data)
}
