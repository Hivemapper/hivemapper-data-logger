package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"

	_ "modernc.org/sqlite"
)

const defaultGnssMeasurementRateHz = 4

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

func queryGnssMeasurementRate(dbPath string) (int, error) {
	// mode=rw so a missing file is an error instead of being created empty.
	db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=rw&_pragma=query_only(1)&_pragma=busy_timeout(2000)")
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var raw string
	err = db.QueryRow("SELECT value FROM config WHERE key = 'GNSS_MEASUREMENT_RATE_HZ'").Scan(&raw)
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
