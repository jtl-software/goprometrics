package main

import (
	"github.com/prometheus/common/log"
	"goprometrics/src/api"
	"goprometrics/src/store"
)

var Version = "development"

func main() {

	config := api.NewConfig()

	go func() {
		metrics := api.NewAdapter(config.MetricsApiHostConfig)
		metrics.ServeMetrics()
	}()

	log.Info("GoPrometrics ", Version)
	adapter := api.NewAdapter(config.ApiHostConfig)

	adapter.CounterHandleFunc(adapter.RequestHandler(store.NewCounterStore()))
	adapter.SummaryHandleFunc(adapter.RequestHandler(store.NewSummaryStore()))
	adapter.HistogramHandleFunc(adapter.RequestHandler(store.NewHistogramStore()))
	adapter.GaugeHandleFunc(adapter.RequestHandler(store.NewGaugeStore()))

	adapter.Serve()
}
