package logger

import "time"

type Logger interface {
	Log(data *Data) error
}

type Data struct {
	SystemTime time.Time `json:"systemtime"`
	Fix        string    `json:"fix"`
	Timestamp  time.Time `json:"timestamp"`

	Latitude   float64     `json:"latitude"`
	Longitude  float64     `json:"longitude"`
	Altitude   float64     `json:"height"`
	Heading    float64     `json:"heading"`
	Speed      float64     `json:"speed"`
	Dop        *Dop        `json:"dop"`
	Satellites *Satellites `json:"satellites"`
	Sep        float64     `json:"sep"` // Estimated Spherical (3D) Position Error in meters. Guessed to be 95% confidence, but many GNSS receivers do not specify, so certainty unknown.
	Eph        float64     `json:"eph"` // Estimated horizontal Position (2D) Error in meters. Also known as Estimated Position Error (epe). Certainty unknown.
}

type Dop struct {
	GDop float64 `json:"gdop"`
	HDop float64 `json:"hdop"`
	PDop float64 `json:"pdop"`
	TDop float64 `json:"tdop"`
	VDop float64 `json:"vdop"`
	XDop float64 `json:"xdop"`
	YDop float64 `json:"ydop"`
}

type Satellites struct {
	Seen int `json:"seen"`
	Used int `json:"used"`
}
