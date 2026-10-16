package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
}

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

// AidingMessage builds the UTC assistance to send at startup. The polynomial terms are
// zero, which makes UTC exactly GPS minus the leap second; the real terms are sub-
// microsecond and arrive with the broadcast. DtLSF equals DtLS, so no leap second is
// advertised as pending.
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

// HandleUbxMessage records the leap second whenever the receiver reports a new one.
func (s *LeapSecondStore) HandleUbxMessage(message interface{}) error {
	timegps, ok := message.(*ubx.NavTimegps)
	if !ok {
		return nil
	}
	if timegps.Valid&ubx.NavTimegpsLeapSValid == 0 {
		return nil
	}
	state := LeapSecondState{LeapSeconds: timegps.LeapS_s, GpsWeek: timegps.Week}
	if state == s.state {
		return nil
	}
	if err := s.save(state); err != nil {
		return fmt.Errorf("saving leap second: %w", err)
	}
	fmt.Printf("Stored GPS leap second %d (week %d) in %s\n",
		state.LeapSeconds, state.GpsWeek, s.path)
	s.state = state
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
