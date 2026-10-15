package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
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
	fallback := defaultGnssConstellations()

	cases := []struct {
		name  string
		value *string
		want  map[string]bool
	}{
		{"missing key falls back", nil, fallback},
		// What the configurator seeds, so this pins the two defaults to agree.
		{"the seeded default set",
			str(`{"BDS":true,"GAL":true,"GPS":true,"QZSS":false,"SBAS":true}`), fallback},
		{"one constellation", str(`{"GPS":true}`), map[string]bool{"GPS": true}},
		{"explicit false is kept", str(`{"GPS":true,"GAL":false}`),
			map[string]bool{"GPS": true, "GAL": false}},
		{"QZSS is supported", str(`{"GPS":true,"QZSS":true}`),
			map[string]bool{"GPS": true, "QZSS": true}},
		{"GLONASS is rejected", str(`{"GPS":true,"GLONASS":true}`), fallback},
		{"unknown name is rejected", str(`{"GPS":true,"NAVIC":true}`), fallback},
		{"lower case is not a name", str(`{"gps":true}`), fallback},
		{"augmentation alone is rejected", str(`{"SBAS":true,"QZSS":true}`), fallback},
		{"all disabled is rejected", str(`{"GPS":false,"GAL":false,"BDS":false}`), fallback},
		{"empty object is rejected", str(`{}`), fallback},
		{"a string is rejected", str(`"GPS,GAL"`), fallback},
		{"a number is rejected", str(`4`), fallback},
		{"malformed json is rejected", str(`{"GPS":`), fallback},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := readGnssConstellations(writeConstellationDb(t, c.value))
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestReadGnssConstellationsMissingDb(t *testing.T) {
	path := filepath.Join(t.TempDir(), "absent.db")
	if got := readGnssConstellations(path); !reflect.DeepEqual(got, defaultGnssConstellations()) {
		t.Fatalf("got %v, want the default set", got)
	}
}
