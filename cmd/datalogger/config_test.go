package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfigDb(t *testing.T, value *string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "odc-api.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE config (key TEXT PRIMARY KEY, value TEXT)"); err != nil {
		t.Fatal(err)
	}
	if value != nil {
		if _, err := db.Exec("INSERT INTO config VALUES ('GNSS_MEASUREMENT_RATE_HZ', ?)", *value); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestReadGnssMeasurementRate(t *testing.T) {
	str := func(s string) *string { return &s }
	cases := []struct {
		name  string
		value *string
		want  int
	}{
		{"configured", str("8"), 8},
		{"lower bound", str("1"), 1},
		{"upper bound", str("25"), 25},
		{"integral float", str("8.0"), 8},
		{"missing key", nil, 4},
		{"zero", str("0"), 4},
		{"too high", str("26"), 4},
		{"fractional", str("8.5"), 4},
		{"json string", str(`"8"`), 4},
		{"not json", str("fast"), 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := readGnssMeasurementRate(writeConfigDb(t, c.value)); got != c.want {
				t.Errorf("got %d, want %d", got, c.want)
			}
		})
	}
}

func TestReadGnssMeasurementRateMissingDb(t *testing.T) {
	path := filepath.Join(t.TempDir(), "odc-api.db")
	if got := readGnssMeasurementRate(path); got != 4 {
		t.Errorf("got %d, want 4", got)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("reading config created %s", path)
	}
}

func TestReadGnssMeasurementRateMissingTable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "odc-api.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("CREATE TABLE other (x)"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	if got := readGnssMeasurementRate(path); got != 4 {
		t.Errorf("got %d, want 4", got)
	}
}

func writeConstellationDb(t *testing.T, value *string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "odc-api.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE config (key TEXT PRIMARY KEY, value TEXT)"); err != nil {
		t.Fatal(err)
	}
	if value != nil {
		if _, err := db.Exec("INSERT INTO config VALUES ('GNSS_CONSTELLATIONS', ?)", *value); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestReadGnssConstellations(t *testing.T) {
	str := func(s string) *string { return &s }
	fallback := "GPS,SBAS,GAL,BDS"

	cases := []struct {
		name  string
		value *string
		want  string
	}{
		{"missing key falls back", nil, fallback},
		{"default set", str(`"GPS,SBAS,GAL,BDS"`), "GPS,SBAS,GAL,BDS"},
		{"single constellation", str(`"GPS"`), "GPS"},
		{"long names and spacing", str(`" galileo , BeiDou "`), "GAL,BDS"},
		{"duplicates collapse", str(`"GPS,gps,GPS"`), "GPS"},
		{"order is preserved", str(`"BDS,GPS"`), "BDS,GPS"},
		{"GLONASS is rejected", str(`"GPS,GLONASS"`), fallback},
		{"unknown name is rejected", str(`"GPS,NAVIC"`), fallback},
		{"augmentation alone is rejected", str(`"SBAS,QZSS"`), fallback},
		{"empty is rejected", str(`""`), fallback},
		{"non-string is rejected", str(`4`), fallback},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := strings.Join(readGnssConstellations(writeConstellationDb(t, c.value)), ",")
			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestReadGnssConstellationsMissingDb(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.db")
	if got := strings.Join(readGnssConstellations(path), ","); got != "GPS,SBAS,GAL,BDS" {
		t.Fatalf("got %q, want the default set", got)
	}
}
