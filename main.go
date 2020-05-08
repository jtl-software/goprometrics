package main

import (
	"goprometrics/src/api"
	"goprometrics/src/store"
)

func main() {

	config := api.NewConfig()

	go func() {
		metrics := api.NewAdapter(config.MetricsApiHostConfig)
		metrics.ServeMetrics()
	}()

	adapter := api.NewAdapter(config.ApiHostConfig)

	adapter.CounterHandleFunc(adapter.RequestHandler(store.NewCounterStore()))
	adapter.SummaryHandleFunc(adapter.RequestHandler(store.NewSummaryStore()))
	adapter.HistogramHandleFunc(adapter.RequestHandler(store.NewHistogramStore()))
	adapter.GaugeHandleFunc(adapter.RequestHandler(store.NewGaugeStore()))

	adapter.Serve()
}
