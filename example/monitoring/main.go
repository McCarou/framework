package main

import (
	"net/http"

	"github.com/radianteam/framework"
	"github.com/radianteam/framework/worker/service/monitoring"
	"github.com/radianteam/framework/worker/service/rest"
)

type MainHandler struct {
	rest.RestServiceHandler
}

// REST handler function
func (h *MainHandler) Handle() error {
	// return standard gin results
	h.GinContext.String(http.StatusOK, "Hello world!\n")

	return nil
}

func main() {
	// create a new framework instance
	radian := framework.NewRadianMicroservice("main")

	// create a new REST worker
	workerConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8088}
	workerRest := rest.NewRestServiceWorker("service_rest", workerConfig)

	// create a new mnitoring worker
	prometheusConfig := &monitoring.MonitoringServiceConfig{Listen: "0.0.0.0", Port: 8087}
	workerPrometheus := monitoring.NewMonitoringServiceWorker("service_monitoring", prometheusConfig)

	// create a route to the worker
	workerRest.SetRoute("GET", "/", &MainHandler{})

	// enable metrics collecting
	workerRest.SetMonitoring(true)

	// append workers to the framework
	radian.AddWorker(workerRest)
	radian.AddWorker(workerPrometheus)

	// run the workers
	radian.RunAll()
}
