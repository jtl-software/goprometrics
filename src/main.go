package main

type ConstLabel struct {
	Name  []string
	Value []string
}

type Counter interface {
	inc(ns string, name string, label ConstLabel, step float64) (bool, error)
}

type PrometheusMetrics struct {
	counterStore CounterStore
}

func main() {
	counter := PrometheusMetrics{counterStore: CounterStore{}}

	go func() {
		NewAdapter().ServeMetrics()
	}()

	api := NewAdapter()
	handler := api.MakeCounterHandler(counter)
	api.CounterHandleFunc(handler)
	api.Serve()
}

func (pm PrometheusMetrics) inc(ns string, name string, label ConstLabel, step float64) (bool, error) {
	created, err := pm.counterStore.addCounter(ns, name, label)
	if err == nil {
		pm.counterStore.inc(ns, name, label, step)
	}
	return created, err
}
