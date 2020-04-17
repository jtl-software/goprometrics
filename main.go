package main

func main() {

	config := NewConfig()
	store := NewMetricStore()

	go func() {
		NewAdapter(config.mApiHostConfig).ServeMetrics()
	}()

	api := NewAdapter(config.apiHostConfig)

	// api
	api.CounterHandleFunc(api.RequestHandler(store, CreateCounterMetricHandler(), IncCounterHandler()))
	api.SummaryHandleFunc(api.RequestHandler(store, CreateSummaryMetricHandler(), ObserveSummaryHandler()))
	api.HistogramHandleFunc(api.RequestHandler(store, CreateHistogramMetricHandler(), ObserveHistogramHandler()))

	api.Serve()
}
