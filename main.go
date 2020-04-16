package main

type ConstLabel struct {
	Name  []string
	Value []string
}

type PrometheusCounter interface {
	incCounter(ns string, name string, label ConstLabel, step float64, help string) (bool, error)
}

type PrometheusSummary interface {
	sum(ns string, name string, label ConstLabel, observation float64, objectives map[float64]float64, help string) (bool, error)
}

type PrometheusHistogram interface {
	observe(ns string, name string, label ConstLabel, observation float64, buckets []float64, help string) (bool, error)
}

type PrometheusMetrics struct {
	counterStore   CounterStore
	summaryStore   SummaryStore
	histogramStore HistogramStore
}

func main() {

	config := NewConfig()
	prometheusMetrics := PrometheusMetrics{
		counterStore:   CounterStore{},
		summaryStore:   SummaryStore{},
		histogramStore: HistogramStore{},
	}

	go func() {
		NewAdapter(config.mApiHostConfig).ServeMetrics()
	}()

	api := NewAdapter(config.apiHostConfig)

	// api
	api.CounterHandleFunc(api.MakeCounterHandler(prometheusMetrics))
	api.SummaryHandleFunc(api.MakeSummaryHandler(prometheusMetrics))
	api.HistogramHandleFunc(api.MakeHistogramHandler(prometheusMetrics))

	api.Serve()
}

func (pm PrometheusMetrics) incCounter(ns string, name string, label ConstLabel, step float64, help string) (bool, error) {
	created, err := pm.counterStore.addCounter(ns, name, label, help)
	if err == nil {
		pm.counterStore.inc(ns, name, label, step)
	}
	return created, err
}

func (pm PrometheusMetrics) sum(ns string, name string, label ConstLabel, observation float64, objectives map[float64]float64, help string) (bool, error) {
	created, err := pm.summaryStore.addSummary(ns, name, label, objectives, help)
	if err == nil {
		pm.summaryStore.observe(ns, name, label, observation)
	}
	return created, err
}

func (pm PrometheusMetrics) observe(ns string, name string, label ConstLabel, observation float64, buckets []float64, help string) (bool, error) {
	created, err := pm.histogramStore.addHistogram(ns, name, label, buckets, help)
	if err == nil {
		pm.histogramStore.observe(ns, name, label, observation)
	}
	return created, err
}
