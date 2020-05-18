package main

import (
	"github.com/prometheus/common/log"
	"goprometrics/src/api"
	"goprometrics/src/store"
	"time"
)

var uniqueCounterStore = store.NewUniqueCounterStore()

func main() {

	config := api.NewConfig()

	go func() {
		metrics := api.NewAdapter(config.MetricsApiHostConfig)
		metrics.ServeMetrics()
	}()

	go gc()

	adapter := api.NewAdapter(config.ApiHostConfig)

	adapter.CounterHandleFunc(adapter.RequestHandler(store.NewCounterStore()))
	adapter.UniqueCounterHandleFunc(adapter.RequestHandler(uniqueCounterStore))
	adapter.SummaryHandleFunc(adapter.RequestHandler(store.NewSummaryStore()))
	adapter.HistogramHandleFunc(adapter.RequestHandler(store.NewHistogramStore()))
	adapter.GaugeHandleFunc(adapter.RequestHandler(store.NewGaugeStore()))

	adapter.Serve()
}

var lastGcRun = time.Now().Unix()

func gc() {

	for {

		timeNow := time.Now().Unix()
		if timeNow-lastGcRun > 60 {
			lastGcRun = timeNow
			for key, uc := range uniqueCounterStore.GetStore() {
				d, l := uc.Gc()
				log.Infof("Collect garbage from unique counter %s with length %d - %d deduplication hashes removed", key, l, d)
			}
		}
		time.Sleep(60)
	}
}
