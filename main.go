package main

import (
	"log/slog"
	"os"

	"goprometrics/src/api"
	"goprometrics/src/store"
)

var Version = "development"

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	config := api.NewConfig()

	go func() {
		metrics := api.NewAdapter(config.MetricsApiHostConfig)
		metrics.ServeMetrics()
	}()

	slog.Info("GoProMetrics started", "version", Version)
	adapter := api.NewAdapter(config.ApiHostConfig)

	adapter.CounterHandleFunc(adapter.RequestHandler(store.NewCounterStore()))
	adapter.SummaryHandleFunc(adapter.RequestHandler(store.NewSummaryStore()))
	adapter.HistogramHandleFunc(adapter.RequestHandler(store.NewHistogramStore()))
	adapter.GaugeHandleFunc(adapter.RequestHandler(store.NewGaugeStore()))

	adapter.Serve()
}
