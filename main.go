package main

import (
	"goprometrics/src/api"
	"goprometrics/src/store"
	"os"
	"os/signal"
)

func main() {

	config := api.NewConfig()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		metrics := api.NewAdapter(config.MetricsApiHostConfig)
		metrics.ServeMetrics()
	}()

	adapter := api.NewAdapter(config.ApiHostConfig)

	adapter.CounterHandleFunc(adapter.RequestHandler(store.NewCounterStore()))
	adapter.SummaryHandleFunc(adapter.RequestHandler(store.NewSummaryStore()))
	adapter.HistogramHandleFunc(adapter.RequestHandler(store.NewHistogramStore()))

	adapter.Serve()
}
