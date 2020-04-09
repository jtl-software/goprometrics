package main

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net/http"
)


func main() {
	Store = CounterStore{}

	go func() {
		m := mux.NewRouter()
		log.Infoln("metrics are getting exposed on :9112")
		m.Path("/").Handler(promhttp.Handler())
		_ = http.ListenAndServe(":9112", m)
	}()

	r := mux.NewRouter()
	log.Infoln("Start Server on :9111")
	r.HandleFunc("/counter/{name}/inc", incCounter).Methods("PUT")
//	r.HandleFunc("/counter/{name}/inc/{labels}").Methods("PUT")

	// delete does not work -
	r.HandleFunc("/prune", pruneCounter).Methods("DELETE")
	_ = http.ListenAndServe(":9111", r)
}

func incCounterVec(writer http.ResponseWriter, request *http.Request) {

}

func pruneCounter(writer http.ResponseWriter, _ *http.Request) {
	log.Infoln("All counters pruned")
	for k := range Store {
		delete(Store, k)
	}
	writer.WriteHeader(http.StatusNoContent)
}

func incCounter(writer http.ResponseWriter, request *http.Request) {

	v := mux.Vars(request)
	name := v["name"]

	status := http.StatusOK
	if _, ok := Store[name]; !ok {
		log.Infof("New counter %s registered", name)

		opts, ok := createNewCounterOpts(name, request)
		if !ok {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		status = http.StatusCreated
		Store[name] = promauto.NewCounter(opts)

		// we have to go with counterVec
		// promauto.NewCounterVec()
	}

	Store[name].Inc()
	writer.WriteHeader(status)
}

