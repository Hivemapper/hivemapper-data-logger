package main

import (
	"database/sql"
	"os"
	"path/filepath"
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
		{"upper bound", str("10"), 10},
		{"integral float", str("8.0"), 8},
		{"missing key", nil, 4},
		{"zero", str("0"), 4},
		{"too high", str("11"), 4},
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
