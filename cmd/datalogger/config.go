package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"

	_ "modernc.org/sqlite"
)

const defaultGnssMeasurementRateHz = 4

// The constellations this receiver can be told about. GLONASS is absent because the
// receiver has no key for it and rejects one.
var supportedConstellations = map[string]bool{
	"GPS": true, "SBAS": true, "GAL": true, "BDS": true, "QZSS": true,
}

// The set every field device already has stored, so the default write is a no-op.
func defaultGnssConstellations() map[string]bool {
	return map[string]bool{"GPS": true, "SBAS": true, "GAL": true, "BDS": true}
}

// readGnssMeasurementRate reads GNSS_MEASUREMENT_RATE_HZ from the odc-api config table.
// bee-sensor-fusion applies the same rule, so both agree on the rate.
func readGnssMeasurementRate(dbPath string) int {
	rate, err := queryGnssMeasurementRate(dbPath)
	if err != nil {
		fmt.Printf("GNSS_MEASUREMENT_RATE_HZ: %v, using %d\n", err, defaultGnssMeasurementRateHz)
		return defaultGnssMeasurementRateHz
	}
	fmt.Printf("GNSS_MEASUREMENT_RATE_HZ: %d\n", rate)
	return rate
}

// readGnssConstellations reads GNSS_CONSTELLATIONS from the odc-api config table.
// Changing the set costs the receiver a ~400 ms loss of fix; writing back the set it
// already holds costs nothing.
func readGnssConstellations(dbPath string) map[string]bool {
	enabled, err := queryGnssConstellations(dbPath)
	if err != nil {
		fmt.Printf("GNSS_CONSTELLATIONS: %v, using %v\n", err, defaultGnssConstellations())
		return defaultGnssConstellations()
	}
	fmt.Printf("GNSS_CONSTELLATIONS: %v\n", enabled)
	return enabled
}

func queryGnssConstellations(dbPath string) (map[string]bool, error) {
	raw, err := queryConfig(dbPath, "GNSS_CONSTELLATIONS")
	if err != nil {
		return nil, err
	}

	var enabled map[string]bool
	if err := json.Unmarshal([]byte(raw), &enabled); err != nil {
		return nil, fmt.Errorf("parsing %q: %w", raw, err)
	}

	for name := range enabled {
		if !supportedConstellations[name] {
			return nil, fmt.Errorf("%q is not a constellation this receiver supports", name)
		}
	}
	// The receiver requires at least one major constellation; SBAS is augmentation and
	// QZSS is regional, so neither counts on its own.
	if !enabled["GPS"] && !enabled["GAL"] && !enabled["BDS"] {
		return nil, fmt.Errorf("%s enables no major constellation (GPS, GAL or BDS)", raw)
	}
	return enabled, nil
}

func queryConfig(dbPath, key string) (string, error) {
	// mode=rw so a missing file is an error instead of being created empty.
	db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=rw&_pragma=query_only(1)&_pragma=busy_timeout(2000)")
	if err != nil {
		return "", err
	}
	defer db.Close()

	var raw string
	if err := db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&raw); err != nil {
		return "", err
	}
	return raw, nil
}

func queryGnssMeasurementRate(dbPath string) (int, error) {
	raw, err := queryConfig(dbPath, "GNSS_MEASUREMENT_RATE_HZ")
	if err != nil {
		return 0, err
	}

	// The configurator serializes every value as JSON text.
	var rate float64
	if err := json.Unmarshal([]byte(raw), &rate); err != nil {
		return 0, fmt.Errorf("parsing %q: %w", raw, err)
	}
	if rate < 1 || rate > 25 || rate != math.Trunc(rate) {
		return 0, fmt.Errorf("%s is not an integer in [1, 25]", raw)
	}
	return int(rate), nil
}
