package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/daedaleanai/ublox/ubx"
)

// The receiver needs the GPS-UTC leap second before it can resolve UTC. It is held in
// battery-backed RAM, which this device has no battery for, so it is lost on every
// power cycle. GPS broadcasts it once per 12.5 minute superframe; Galileo carries it
// every I/NAV subframe, which is why losing Galileo made the wait visible. Until UTC
// resolves, fusion writes no GNSS rows and clips are recorded without position.
//
// The leap second has been 18 s since 2017 and changes are announced months ahead, so
// the last value seen is a safe starting point. The receiver overwrites it with the
// broadcast value as soon as it decodes one.
type LeapSecondStore struct {
	path  string
	state LeapSecondState
}

type LeapSecondState struct {
	LeapSeconds int8  `json:"leap_seconds"`
	GpsWeek     int16 `json:"gps_week"`
	// UTC at the moment the leap second was stored, derived from GPS week and
	// time-of-week rather than the system clock, which is stale at boot.
	UtcUnix int64 `json:"utc_unix"`
}

// gpsEpoch is 1980-01-06 00:00:00 UTC.
var gpsEpoch = time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)

func NewLeapSecondStore(path string) *LeapSecondStore {
	return &LeapSecondStore{path: path}
}

// Load returns the stored leap second, or false if there is nothing usable on disk.
func (s *LeapSecondStore) Load() (LeapSecondState, bool) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("reading leap second file:", err)
		}
		return LeapSecondState{}, false
	}
	var state LeapSecondState
	if err := json.Unmarshal(data, &state); err != nil {
		fmt.Printf("parsing leap second file %s: %v\n", s.path, err)
		return LeapSecondState{}, false
	}
	// A zero leap second is not a value this receiver can legitimately report.
	if state.LeapSeconds == 0 {
		return LeapSecondState{}, false
	}
	s.state = state
	return state, true
}

// TimeMessage carries the stored UTC. The receiver rejects UTC assistance while it has
// no time of its own (UBX-MGA-ACK infoCode 1), and this is what the ack text says to
// send first. There is no RTC, so the stored time is stale by however long the device
// was off; the accuracy field is set to its maximum to say so rather than claim a
// precision we do not have.
func TimeMessage(state LeapSecondState) *ubx.MgaIniTime_utc6 {
	t := time.Unix(state.UtcUnix, 0).UTC()
	return &ubx.MgaIniTime_utc6{
		Type:       0x10,
		Version:    0x00,
		LeapSecs_s: state.LeapSeconds,
		Year:       uint16(t.Year()),
		Month:      byte(t.Month()),
		Day:        byte(t.Day()),
		Hour:       byte(t.Hour()),
		Minute:     byte(t.Minute()),
		Second_s:   byte(t.Second()),
		TAccS_s:    ^uint16(0),
	}
}

// AidingMessage builds the UTC assistance. The polynomial terms are zero, which makes
// UTC exactly GPS minus the leap second; the real terms are sub-microsecond and arrive
// with the broadcast. DtLSF equals DtLS, so no leap second is advertised as pending.
func AidingMessage(state LeapSecondState) *ubx.MgaGpsUtc4 {
	return &ubx.MgaGpsUtc4{
		Type:           0x05,
		Version:        0x00,
		UtcDtLS_s:      state.LeapSeconds,
		UtcDtLSF_s:     state.LeapSeconds,
		UtcWNt_weeks:   byte(state.GpsWeek % 256),
		UtcWNlsf_weeks: byte(state.GpsWeek % 256),
		UtcDn_days:     7,
	}
}

var mgaInfoCodes = map[byte]string{
	0: "accepted",
	1: "receiver does not know the time",
	2: "message version not supported",
	3: "message size does not match version",
	4: "could not be stored to the database",
	5: "receiver not ready",
	6: "unknown message type",
}

// HandleUbxMessage records the leap second the receiver reports, and reports what it
// did with the assistance sent at startup. Without the ack an ignored message looks
// exactly like one that worked.
func (s *LeapSecondStore) HandleUbxMessage(message interface{}) error {
	switch m := message.(type) {
	case *ubx.MgaAckData0:
		// 0x13 0x40 is MGA-INI-TIME, 0x13 0x00 is MGA-GPS-UTC. Other MGA traffic is
		// the AssistNow Offline load, which counts its own acks.
		name, ours := map[byte]string{0x40: "MGA-INI-TIME_UTC", 0x00: "MGA-GPS-UTC"}[m.MsgId]
		if !ours {
			return nil
		}
		verdict := "USED"
		if m.Type == 0 {
			verdict = "NOT USED"
		}
		fmt.Printf("%s: %s (infoCode %d: %s)\n",
			name, verdict, m.InfoCode, mgaInfoCodes[m.InfoCode])
	case *ubx.NavTimegps:
		if m.Valid&ubx.NavTimegpsLeapSValid == 0 {
			return nil
		}
		utc := gpsEpoch.
			Add(time.Duration(m.Week) * 7 * 24 * time.Hour).
			Add(time.Duration(m.ITOW_ms) * time.Millisecond).
			Add(-time.Duration(m.LeapS_s) * time.Second)
		state := LeapSecondState{
			LeapSeconds: m.LeapS_s,
			GpsWeek:     m.Week,
			UtcUnix:     utc.Unix(),
		}
		// The time moves constantly; only a new leap second is worth a write.
		if state.LeapSeconds == s.state.LeapSeconds && state.GpsWeek == s.state.GpsWeek {
			return nil
		}
		if err := s.save(state); err != nil {
			return fmt.Errorf("saving leap second: %w", err)
		}
		fmt.Printf("Stored GPS leap second %d (week %d, utc %s) in %s\n",
			state.LeapSeconds, state.GpsWeek, utc.Format(time.RFC3339), s.path)
		s.state = state
	}
	return nil
}

// save writes via a temp file and rename, so an interrupted write cannot leave a
// truncated file to be read back on the next boot.
func (s *LeapSecondStore) save(state LeapSecondState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
