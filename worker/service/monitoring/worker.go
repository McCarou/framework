package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/radianteam/framework/worker"
)

type MonitoringServiceConfig struct {
	Listen string `json:"Listen,omitempty" config:"Listen,required"`
	Port   int16  `json:"Port,omitempty" config:"Port,required"`
}

type MonitoringServiceWorker struct {
	*worker.BaseWorker

	config *MonitoringServiceConfig

	server *http.Server
}

func NewMonitoringServiceWorker(name string, config *MonitoringServiceConfig) *MonitoringServiceWorker {
	return &MonitoringServiceWorker{
		BaseWorker: worker.NewBaseWorker(name),
		config:     config,
	}
}

func (w *MonitoringServiceWorker) Setup() {
	w.Logger.Infof("Setting up monitoring Service")

	srvMux := http.NewServeMux()
	srvMux.Handle("/metrics", promhttp.Handler())

	w.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", w.config.Listen, w.config.Port),
		Handler: srvMux,
	}
}

func (w *MonitoringServiceWorker) Run() {
	w.Logger.Infof("Running monitoring Service")

	if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		w.Logger.Fatalf("listen %s\n", err)
	}
}

func (w *MonitoringServiceWorker) Stop() {
	w.Logger.Infof("stop signal received! Graceful shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.server.Shutdown(ctx); err != nil {
		w.Logger.Fatal("Server forced to shutdown: ", err)
	}
}
