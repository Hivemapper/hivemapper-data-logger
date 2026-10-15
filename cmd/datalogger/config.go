package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	_ "modernc.org/sqlite"
)

const defaultGnssMeasurementRateHz = 4

// The set every field device already has stored, so the default write is a no-op.
// GLONASS is absent: the receiver has no key for it and rejects one.
const defaultGnssConstellations = "GPS,SBAS,GAL,BDS"

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

// readGnssConstellations reads GNSS_CONSTELLATIONS from the odc-api config table
// and returns the constellations to enable. Changing the set costs the receiver a
// ~400 ms loss of fix; writing back the set it already holds costs nothing.
func readGnssConstellations(dbPath string) []string {
	names, err := queryGnssConstellations(dbPath)
	if err != nil {
		fmt.Printf("GNSS_CONSTELLATIONS: %v, using %s\n", err, defaultGnssConstellations)
		return parseConstellations(defaultGnssConstellations)
	}
	fmt.Printf("GNSS_CONSTELLATIONS: %s\n", strings.Join(names, ","))
	return names
}

func queryGnssConstellations(dbPath string) ([]string, error) {
	raw, err := queryConfig(dbPath, "GNSS_CONSTELLATIONS")
	if err != nil {
		return nil, err
	}
	var value string
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf("parsing %q: %w", raw, err)
	}
	return validateConstellations(value)
}

func parseConstellations(value string) []string {
	names, err := validateConstellations(value)
	if err != nil {
		panic(err) // only reached if defaultGnssConstellations is itself invalid
	}
	return names
}

// validateConstellations normalises the list and rejects anything the receiver
// cannot do, so a typo falls back to the default instead of being written.
func validateConstellations(value string) ([]string, error) {
	supported := map[string]string{
		"GPS": "GPS", "SBAS": "SBAS",
		"GAL": "GAL", "GALILEO": "GAL",
		"BDS": "BDS", "BEIDOU": "BDS",
		"QZSS": "QZSS",
	}
	seen := map[string]bool{}
	names := []string{}
	for _, field := range strings.Split(value, ",") {
		field = strings.ToUpper(strings.TrimSpace(field))
		if field == "" {
			continue
		}
		name, ok := supported[field]
		if !ok {
			return nil, fmt.Errorf("%q is not a constellation this receiver supports", field)
		}
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	// The receiver requires at least one major constellation; SBAS is augmentation
	// and QZSS is regional, so neither counts on its own.
	if !seen["GPS"] && !seen["GAL"] && !seen["BDS"] {
		return nil, fmt.Errorf("%q enables no major constellation (GPS, GAL or BDS)", value)
	}
	return names, nil
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
