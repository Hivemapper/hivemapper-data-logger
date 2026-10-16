package handlers

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/daedaleanai/ublox/ubx"
)

func store(t *testing.T) (*LeapSecondStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gnss-leap-second.json")
	return NewLeapSecondStore(path), path
}

func timegps(leap int8, week int16, valid ubx.NavTimegpsValid) *ubx.NavTimegps {
	return &ubx.NavTimegps{LeapS_s: leap, Week: week, Valid: valid}
}

func TestLoadWithNothingStored(t *testing.T) {
	s, _ := store(t)
	if _, ok := s.Load(); ok {
		t.Fatal("expected no stored leap second")
	}
}

func TestRoundTrip(t *testing.T) {
	s, path := store(t)
	if err := s.HandleUbxMessage(timegps(18, 2400, ubx.NavTimegpsLeapSValid)); err != nil {
		t.Fatal(err)
	}

	state, ok := NewLeapSecondStore(path).Load()
	if !ok {
		t.Fatal("expected a stored leap second")
	}
	if state.LeapSeconds != 18 || state.GpsWeek != 2400 {
		t.Fatalf("got %+v, want leap 18 week 2400", state)
	}
}

func TestIgnoresMessagesWithoutTheValidFlag(t *testing.T) {
	s, path := store(t)
	// Tow and week valid, leap second not.
	if err := s.HandleUbxMessage(timegps(18, 2400,
		ubx.NavTimegpsTowValid|ubx.NavTimegpsWeekValid)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("wrote a file from a message with no valid leap second")
	}
}

func TestIgnoresOtherMessageTypes(t *testing.T) {
	s, path := store(t)
	if err := s.HandleUbxMessage(&ubx.NavPvt{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("wrote a file from an unrelated message")
	}
}

func TestRewritesOnlyWhenTheValueChanges(t *testing.T) {
	s, path := store(t)
	if err := s.HandleUbxMessage(timegps(18, 2400, ubx.NavTimegpsLeapSValid)); err != nil {
		t.Fatal(err)
	}
	first, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	// Same value again: nothing should be written back.
	if err := s.HandleUbxMessage(timegps(18, 2400, ubx.NavTimegpsLeapSValid)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("rewrote the file for an unchanged value")
	}

	// A new leap second must be persisted.
	if err := s.HandleUbxMessage(timegps(19, 2401, ubx.NavTimegpsLeapSValid)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("did not persist a changed leap second: %v (first write %v)", err, first.Size())
	}
	state, _ := NewLeapSecondStore(path).Load()
	if state.LeapSeconds != 19 {
		t.Fatalf("got %d, want 19", state.LeapSeconds)
	}
}

func TestRejectsGarbageAndZero(t *testing.T) {
	for _, c := range []struct{ name, body string }{
		{"not json", "this is not json"},
		{"truncated", `{"leap_seconds":`},
		{"zero leap seconds", `{"leap_seconds":0,"gps_week":2400}`},
	} {
		t.Run(c.name, func(t *testing.T) {
			s, path := store(t)
			if err := os.WriteFile(path, []byte(c.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, ok := s.Load(); ok {
				t.Fatal("accepted an unusable file")
			}
		})
	}
}

// The receiver silently ignores a malformed frame, so pin the bytes rather than trust
// the struct. 20-byte payload, type 0x05, version 0x00, DtLS and DtLSF both 18.
func TestAidingMessageEncoding(t *testing.T) {
	b, err := ubx.Encode(AidingMessage(LeapSecondState{LeapSeconds: 18, GpsWeek: 2400}))
	if err != nil {
		t.Fatal(err)
	}
	got := hex.EncodeToString(b)
	want := "b56213001400050000000000000000000000120060600712000017e6"
	if got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
}
