package logger

import (
	"fmt"
	"sync"
	"time"
)

var (
	prevItowMs = map[string]uint32{
		"NavPvt":     0,
		"NavCov":     0,
		"NavTimegps": 0,
		"NavPosecef": 0,
		"NavVelecef": 0,
		"NavStatus":  0,
		"NavDop":     0,
		"NavSig":     0,
		"TimTp":      0}
	// GnssMeasurementRateHz is the configured GNSS rate, set at startup.
	GnssMeasurementRateHz = 4

	// Guards the counters below: noteItow runs on the decoder goroutine, the reporter
	// on its own.
	gnssStatsMu    sync.Mutex
	gnssRecvCounts = map[string]int{}
	gnssStatsStart time.Time
	gnssSawData    bool
)

// gnssStatsKeys are the messages fusion needs on one itow to write an epoch.
var gnssStatsKeys = []string{"NavPvt", "NavCov", "NavTimegps", "NavPosecef", "NavVelecef", "NavStatus", "NavDop", "NavSig"}

const gnssStatsInterval = 60 * time.Second

// noteItow counts an arrival. StartGnssStatsReporter prints the totals.
func noteItow(key string, itow uint32) {
	prevItowMs[key] = itow

	gnssStatsMu.Lock()
	defer gnssStatsMu.Unlock()
	gnssRecvCounts[key]++
	if !gnssSawData {
		// Start the window at the first message, not at boot: Init spends seconds
		// waiting on config ACKs and would otherwise count as a shortfall.
		gnssSawData = true
		gnssStatsStart = time.Now()
	}
}

// StartGnssStatsReporter prints per-message arrival counts once a minute. One message low
// means its output rate is wrong; all low means whole epochs are being lost. Driven by a
// timer rather than by arrivals so it still reports when nothing is arriving at all.
func StartGnssStatsReporter() {
	go func() {
		for range time.Tick(gnssStatsInterval) {
			reportGnssStats()
		}
	}()
}

func reportGnssStats() {
	gnssStatsMu.Lock()
	defer gnssStatsMu.Unlock()

	if !gnssSawData {
		fmt.Printf("%v [WARNING] [GNSS] no nav messages received\n", time.Now().UTC())
		return
	}

	elapsed := time.Since(gnssStatsStart)
	expected := int(float64(GnssMeasurementRateHz) * elapsed.Seconds())
	counts := ""
	lowest := expected
	for _, k := range gnssStatsKeys {
		counts += fmt.Sprintf(" %s=%d", k, gnssRecvCounts[k])
		if gnssRecvCounts[k] < lowest {
			lowest = gnssRecvCounts[k]
		}
	}
	tag := "[GNSS]"
	if expected > 0 && lowest*10 < expected*9 {
		tag = "[WARNING] [GNSS]"
	}
	fmt.Printf("%v %s %.0fs: expected %d epochs @%dHz, received%s\n",
		time.Now().UTC(), tag, elapsed.Seconds(), expected, GnssMeasurementRateHz, counts)

	for k := range gnssRecvCounts {
		delete(gnssRecvCounts, k)
	}
	gnssStatsStart = time.Now()
}
