package serviceworker

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/radianteam/framework/worker"

	"github.com/gin-gonic/gin"
)

type RegFuncRestServiceWorker func(c *gin.Context, wc *worker.WorkerAdapters)

type RestConfig struct {
	Listen string `json:"listen,omitempty" config:"listen,required"`
	Port   int16  `json:"port,omitempty" config:"port,required"`
}

type RestServiceWorker struct {
	*worker.BaseWorker

	config *RestConfig

	routes *gin.Engine
	server *http.Server
}

func NewRestServiceWorker(name string, config *RestConfig) *RestServiceWorker {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	return &RestServiceWorker{
		BaseWorker: worker.NewBaseWorker(name),
		config:     config,
		routes:     gin.Default(),
	}
}

func (w *RestServiceWorker) SetRoute(method string, path string, handler RegFuncRestServiceWorker) {
	w.routes.Handle(strings.ToUpper(method), path, func(c *gin.Context) {
		handler(c, w.Adapters)
	})
}

func (w *RestServiceWorker) Setup() {
	w.Logger.Infof("Setting up REST Service")

	w.routes.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := GetDurationInMillseconds(start)

		entry := w.Logger.WithFields(logrus.Fields{
			"client_ip": GetClientIP(c),
			"duration":  duration,
			"method":    c.Request.Method,
			"path":      c.Request.RequestURI,
			"status":    c.Writer.Status(),
		})

		if c.Writer.Status() >= 500 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("Request has been completed")
		}
	})

	w.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", w.config.Listen, w.config.Port),
		Handler: w.routes,
	}
}

func (w *RestServiceWorker) Run() {
	w.Logger.Infof("Running REST Service")

	if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		w.Logger.Fatalf("listen %s\n", err)
	}
}

func (w *RestServiceWorker) Stop() {
	w.Logger.Infof("stop signal received! Graceful shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.server.Shutdown(ctx); err != nil {
		w.Logger.Fatal("Server forced to shutdown: ", err)
	}
}
