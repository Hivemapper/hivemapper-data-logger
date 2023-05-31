package data

import "fmt"

type AverageFloat64 struct {
	name    string
	entries []float64
	sum     float64
	Average float64
}

func NewAverageFloat64(name string) *AverageFloat64 {
	return &AverageFloat64{name: name}
}

func (a *AverageFloat64) Add(value float64) {
	a.entries = append(a.entries, value)
	a.sum += value

	if len(a.entries) == 101 {
		var first float64
		first, a.entries = a.entries[0], a.entries[1:]

		a.sum -= first
		a.Average = a.sum / 100
	}
}

func (a *AverageFloat64) String() string {
	return fmt.Sprintf("%s: %f", a.name, a.Average)
}
