package handlers

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/daedaleanai/ublox/ubx"
)

func store(t *testing.T) (*LeapSecondStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gnss-leap-second.json")
	return NewLeapSecondStore(path), path
}

func timegps(leap int8, week int16, valid ubx.NavTimegpsValid) *ubx.NavTimegps {
	// 3 days into the week, so the derived UTC is not the week boundary.
	return &ubx.NavTimegps{
		LeapS_s: leap, Week: week, Valid: valid,
		ITOW_ms: 3 * 24 * 60 * 60 * 1000,
	}
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
	// GPS week 2400 begins 2026-01-04; +3 days, -18 leap seconds.
	want := time.Date(2026, 1, 6, 23, 59, 42, 0, time.UTC).Unix()
	if state.UtcUnix != want {
		t.Fatalf("got utc %s, want %s",
			time.Unix(state.UtcUnix, 0).UTC(), time.Unix(want, 0).UTC())
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

// The time message must be well formed and must go out before the UTC assistance; the
// receiver rejects UTC data outright while it has no clock (infoCode 1).
func TestTimeMessageEncoding(t *testing.T) {
	b, err := ubx.Encode(TimeMessage(LeapSecondState{
		LeapSeconds: 18,
		GpsWeek:     2400,
		UtcUnix:     time.Date(2026, 1, 6, 23, 59, 42, 0, time.UTC).Unix(),
	}))
	if err != nil {
		t.Fatal(err)
	}
	got := hex.EncodeToString(b)
	if !strings.HasPrefix(got, "b5621340") {
		t.Fatalf("not an MGA-INI-TIME_UTC frame: %s", got)
	}
	p := b[6 : len(b)-2]
	if p[0] != 0x10 || p[1] != 0x00 {
		t.Fatalf("type/version = %#x/%#x, want 0x10/0x00", p[0], p[1])
	}
	if int8(p[3]) != 18 {
		t.Fatalf("leap seconds = %d, want 18", int8(p[3]))
	}
	year := int(p[4]) | int(p[5])<<8
	if year != 2026 || p[6] != 1 || p[7] != 6 {
		t.Fatalf("date = %d-%02d-%02d, want 2026-01-06", year, p[6], p[7])
	}
	// Accuracy is claimed as the widest the field allows: with no RTC we cannot know
	// how stale the stored time is.
	if acc := int(p[16]) | int(p[17])<<8; acc != 0xFFFF {
		t.Fatalf("time accuracy = %d, want 65535", acc)
	}
}

// A rejected assistance message is acknowledged just like an accepted one, so the
// verdict has to be read out of the ack rather than inferred from its arrival.
func TestAckReporting(t *testing.T) {
	s, _ := store(t)
	for _, c := range []struct {
		name  string
		msgID byte
		typ   byte
		info  byte
	}{
		{"utc accepted", 0x00, 1, 0},
		{"utc rejected, no time", 0x00, 0, 1},
		{"time accepted", 0x40, 1, 0},
		{"some other mga message", 0x20, 0, 4},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := s.HandleUbxMessage(&ubx.MgaAckData0{
				Type: c.typ, InfoCode: c.info, MsgId: c.msgID,
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
