package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/radianteam/framework"
	"github.com/radianteam/framework/worker"
	"github.com/radianteam/framework/worker/service/monitoring"
	"github.com/radianteam/framework/worker/service/rest"
)

// REST handler function
func handlerMain(c *gin.Context, wc *worker.WorkerAdapters) {
	// return standard gin results
	c.String(http.StatusOK, "Hello world!\n")
}

func main() {
	// create a new framework instance
	radian := framework.NewRadianFramework()

	// create a new REST worker
	workerConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8088}
	workerRest := rest.NewRestServiceWorker("service_rest", workerConfig)

	// create a new mnitoring worker
	prometheus_config := &monitoring.MonitoringServiceConfig{Listen: "0.0.0.0", Port: 8087}
	worker_prometheus := monitoring.NewMonitoringServiceWorker("service_monitoring", prometheus_config)

	// create a route to the worker
	workerRest.SetRoute("GET", "/", handlerMain)

	workerRest.SetMonitoring(true)

	// append worker to the framework
	radian.AddWorker(workerRest)
	radian.AddWorker(worker_prometheus)

	// run the worker
	radian.Run([]string{workerRest.GetName(), worker_prometheus.GetName()})
}
