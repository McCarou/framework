package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/radianteam/framework/worker"

	"github.com/gin-gonic/gin"
)

type RestConfig struct {
	Listen string `json:"Listen,omitempty" config:"Listen,required"`
	Port   int16  `json:"Port,omitempty" config:"Port,required"`
}

type RestServiceWorker struct {
	*worker.BaseWorker

	config *RestConfig

	routes *gin.Engine
	server *http.Server

	metricRequestCount *prometheus.CounterVec // TODO: refactor to a Gin metric instrumentor
}

func NewRestServiceWorker(name string, config *RestConfig) *RestServiceWorker {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	engine := gin.Default()

	wrkr := &RestServiceWorker{
		BaseWorker: worker.NewBaseWorker(name),
		config:     config,
		routes:     engine,
		metricRequestCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "rest_worker_total_requests",
			Help: "Total requests of the rest worker",
		}, []string{"worker_name", "code", "method", "url"}),
	}

	engine.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := GetDurationInMillseconds(start)

		entry := wrkr.Logger.WithFields(logrus.Fields{
			"client_ip": GetClientIP(c),
			"duration":  duration,
			"method":    c.Request.Method,
			"path":      c.Request.RequestURI,
			"status":    c.Writer.Status(),
		})

		if wrkr.IsMonitoringEnable() { // TODO: refactor to 2 separated meddlawares
			wrkr.metricRequestCount.With(
				prometheus.Labels{
					"worker_name": wrkr.GetName(),
					"code":        fmt.Sprintf("%d", c.Writer.Status()),
					"method":      c.Request.Method,
					"url":         c.Request.RequestURI,
				}).Inc()
		}

		if c.Writer.Status() >= 500 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("Request has been completed")
		}
	})

	return wrkr
}

func (w *RestServiceWorker) SetRoute(method string, path string, handler RestServiceHandlerInterface) {
	handler.SetLogger(w.Logger.WithField("path", fmt.Sprintf("%s %s", method, path))) // TODO: move to setup
	handler.SetAdapters(w.Adapters)                                                   // TODO: move to setup

	w.routes.Handle(strings.ToUpper(method), path, func(c *gin.Context) {
		handler.SetGinContext(c)

		err := handler.Handle()

		if err != nil {
			w.Logger.Errorf("Method %s with a path %s has been completed with an error: %v", method, path, err)

			// TODO: rewrite status if it is still ok
		}
	})
}

func (w *RestServiceWorker) Setup() {
	w.Logger.Info("Setting up REST Service")

	prometheus.MustRegister(w.metricRequestCount) // TODO: refactor to Register

	w.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", w.config.Listen, w.config.Port),
		Handler: w.routes,
	}
}

func (w *RestServiceWorker) Run() {
	w.Logger.Info("Running REST Service")

	if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		w.Logger.Fatalf("listen %s\n", err)
	}
}

func (w *RestServiceWorker) Stop() {
	w.Logger.Info("stop signal received! Graceful shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.server.Shutdown(ctx); err != nil {
		w.Logger.Fatal("Server forced to shutdown: ", err)
	}
}
