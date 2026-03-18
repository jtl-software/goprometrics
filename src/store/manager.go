package store

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var mutex sync.Mutex

var (
	AppendCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "goprometrics",
		Name:      "metric_appended",
		Help:      "Count when a new metric is appended",
	})

	ErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "goprometrics",
		Name:      "metric_appended_error",
		Help:      "Count when there is a error during append a new metric",
	})
)

func Append(s Store, opts MetricOpts) (new bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", v)
			}
			slog.Error("Panic recovered during metric append", "err", err)
			ErrorCounter.Inc()
		}
	}()

	new = false
	if !s.Has(opts) {
		mutex.Lock()
		defer mutex.Unlock()

		if !s.Has(opts) {
			s.Append(opts)
			new = true
			AppendCounter.Inc()
		}
	}
	return
}
