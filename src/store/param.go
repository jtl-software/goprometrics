package store

import "strings"

type ConstLabel struct {
	Name  []string
	Value []string
}

type MetricOpts struct {
	Ns                string
	Name              string
	Label             ConstLabel
	Help              string
	HistogramBuckets  []float64
	SummaryObjectives map[float64]float64
	SetGaugeToValue   bool
}

func (opts MetricOpts) Key() string {
	return opts.Ns + "_" + opts.Name + "__" + strings.Join(opts.Label.Name, "_")
}
