package store

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/log"
	"math/rand"
	"sync"
	"time"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

type (
	Store interface {
		Append(opts MetricOpts)
		Inc(opts MetricOpts, value float64)
		Has(opts MetricOpts) bool
	}
	counterStore struct {
		store map[string]*prometheus.CounterVec
	}
	summaryStore struct {
		store map[string]*prometheus.SummaryVec
	}
	histogramStore struct {
		store map[string]*prometheus.HistogramVec
	}
	gaugeStore struct {
		store map[string]*prometheus.GaugeVec
	}
	UniqueCounter struct {
		keyMap   sync.Map
		counter  *prometheus.CounterVec
		ucounter *prometheus.CounterVec
	}
	UniqueCounterStore struct {
		store map[string]*UniqueCounter
	}
)

func NewCounterStore() Store {
	return counterStore{
		store: map[string]*prometheus.CounterVec{},
	}
}

func NewUniqueCounterStore() UniqueCounterStore {
	return UniqueCounterStore{
		store: map[string]*UniqueCounter{},
	}
}

func NewSummaryStore() Store {
	return summaryStore{
		store: map[string]*prometheus.SummaryVec{},
	}
}

func NewHistogramStore() Store {
	return histogramStore{
		store: map[string]*prometheus.HistogramVec{},
	}
}

func NewGaugeStore() Store {
	return gaugeStore{
		store: map[string]*prometheus.GaugeVec{},
	}
}

//Counter
func (s counterStore) Has(opts MetricOpts) bool {
	_, has := s.store[opts.Key()]
	return has
}

func (s counterStore) Append(opts MetricOpts) {
	s.store[opts.Key()] = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: opts.Ns,
			Name:      opts.Name,
			Help:      opts.Help,
		},
		opts.Label.Name,
	)
	log.Infof("A new counter %s_%s with labels %v registered", opts.Ns, opts.Name, opts.Label.Name)
}

func (s counterStore) Inc(opts MetricOpts, value float64) {
	if s.Has(opts) {
		s.store[opts.Key()].WithLabelValues(opts.Label.Value...).Add(value)
	}
}

//Summary
func (s summaryStore) Has(opts MetricOpts) bool {
	_, has := s.store[opts.Key()]
	return has
}

func (s summaryStore) Append(opts MetricOpts) {
	s.store[opts.Key()] = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  opts.Ns,
			Name:       opts.Name,
			Help:       opts.Help,
			Objectives: opts.SummaryObjectives,
		},
		opts.Label.Name,
	)
	log.Infof("A new summary %s_%s with labels %v and objectives %v registered", opts.Ns, opts.Name, opts.Label.Name, opts.SummaryObjectives)
}

func (s summaryStore) Inc(opts MetricOpts, value float64) {
	if s.Has(opts) {
		s.store[opts.Key()].WithLabelValues(opts.Label.Value...).Observe(value)
	}
}

// Histogram
func (s histogramStore) Has(opts MetricOpts) bool {
	_, has := s.store[opts.Key()]
	return has
}

func (s histogramStore) Append(opts MetricOpts) {
	s.store[opts.Key()] = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: opts.Ns,
			Name:      opts.Name,
			Help:      opts.Help,
			Buckets:   opts.HistogramBuckets,
		},
		opts.Label.Name,
	)
	log.Infof("A new histogram %s_%s with labels %v and buckets %v registered", opts.Ns, opts.Name, opts.Label.Name, opts.HistogramBuckets)
}

func (s histogramStore) Inc(opts MetricOpts, value float64) {
	if s.Has(opts) {
		s.store[opts.Key()].WithLabelValues(opts.Label.Value...).Observe(value)
	}
}

// Gauge
func (g gaugeStore) Append(opts MetricOpts) {
	g.store[opts.Key()] = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: opts.Ns,
			Name:      opts.Name,
			Help:      opts.Help,
		},
		opts.Label.Name,
	)
	log.Infof("A new gauge %s_%s with labels %v registered", opts.Ns, opts.Name, opts.Label.Name)
}

func (g gaugeStore) Inc(opts MetricOpts, value float64) {
	if g.Has(opts) {
		g.store[opts.Key()].WithLabelValues(opts.Label.Value...).Add(value)
	}
}

func (g gaugeStore) Has(opts MetricOpts) bool {
	_, has := g.store[opts.Key()]
	return has
}

// DeDup Counter Store
func (d UniqueCounterStore) Append(opts MetricOpts) {
	d.store[opts.Key()] = &UniqueCounter{
		keyMap: sync.Map{},
		ucounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: opts.Ns,
				Name:      "unq_" + opts.Name,
				Help:      opts.Help,
			},
			opts.Label.Name,
		),
		counter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: opts.Ns,
				Name:      opts.Name,
				Help:      opts.Help,
			},
			opts.Label.Name,
		),
	}
	log.Infof("A new UniqueCounter %s_%s with labels %v registered", opts.Ns, opts.Name, opts.Label.Name)
}

func (d UniqueCounterStore) Inc(opts MetricOpts, value float64) {
	if d.Has(opts) {
		k := opts.Key()
		d.store[k].counter.WithLabelValues(opts.Label.Value...).Add(value)

		hasCounted := d.isCounted(opts, k)
		if hasCounted == false {
			d.store[k].ucounter.WithLabelValues(opts.Label.Value...).Add(value)
		}

		if seededRand.Intn(1000000) == 1 {
			go d.store[k].Gc()
		}
	}
}

func (d UniqueCounterStore) isCounted(opts MetricOpts, k string) bool {
	_, hasCounted := d.store[k].keyMap.LoadOrStore(opts.DedupHash, time.Now().Unix()+(60))
	return hasCounted
}

func (d UniqueCounterStore) Has(opts MetricOpts) bool {
	_, has := d.store[opts.Key()]
	return has
}

func (d UniqueCounterStore) GetStore() map[string]*UniqueCounter {
	return d.store
}

func (d *UniqueCounter) Gc() (deleted int64, length int64) {

	gcTime := time.Now().Unix()
	d.keyMap.Range(func(key, value interface{}) bool {
		if value.(int64) <= gcTime {
			d.keyMap.Delete(key)
			deleted++
		}
		length++
		return true
	})

	return deleted, length
}
