package data

import "fmt"

type AverageFloat64 struct {
	name       string
	entries    []float64
	sum        float64
	Average    float64
	entryCount int
}

func NewAverageFloat64(name string) *AverageFloat64 {
	return &AverageFloat64{name: name, entryCount: 10}
}

func NewAverageFloat64WithCount(name string, entryCount int) *AverageFloat64 {
	return &AverageFloat64{name: name, entryCount: entryCount}
}

func (a *AverageFloat64) Add(value float64) {
	a.entries = append(a.entries, value)
	a.sum += value

	if len(a.entries) == a.entryCount+1 {
		var first float64
		first, a.entries = a.entries[0], a.entries[1:]
		a.sum -= first
	}
	a.Average = a.sum / float64(len(a.entries))
}

func (a *AverageFloat64) Reset() {
	a.entries = nil
	a.sum = 0
	a.Average = 0
}

func (a *AverageFloat64) String() string {
	return fmt.Sprintf("%s: %f", a.name, a.Average)
}
