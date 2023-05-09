package handlers

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/daedaleanai/ublox/ubx"
)

type TimeGetter struct {
	done chan time.Time
}

func NewTimeGetter(done chan time.Time) *TimeGetter {
	return &TimeGetter{done: done}
}

func (g *TimeGetter) HandleUbxMessage(message interface{}) error {
	navPvt := message.(*ubx.NavPvt)
	fmt.Println("time getter nav pvt info, date validity:", navPvt.Valid, "accuracy:", navPvt.TAcc_ns, "lock type:", navPvt.FixType, "flags:", navPvt.Flags, "flags2:", navPvt.Flags2, "flags3:", navPvt.Flags3)
	if navPvt.Valid&0x1 == 0 {
		return nil
	}
	now := time.Date(int(navPvt.Year_y), time.Month(int(navPvt.Month_month)), int(navPvt.Day_d), int(navPvt.Hour_h), int(navPvt.Min_min), int(navPvt.Sec_s), int(navPvt.Nano_ns), time.UTC)
	fmt.Println("Got a valid date:", now)

	g.done <- now
	return nil
}

func SetSystemDate(newTime time.Time) error {
	_, err := exec.LookPath("date")
	if err != nil {
		return fmt.Errorf("look for date binary: %w", err)
	} else {
		dateString := newTime.Format("2006-01-02 15:04:05")
		//dateString := newTime.Format("2 Jan 2006 15:04:05")
		fmt.Printf("Setting system date to: %s\n", dateString)
		args := []string{"--set", dateString}
		cmd := exec.Command("date", args...)
		fmt.Println("Running cmd:", cmd.String())
		return cmd.Run()
	}
}
