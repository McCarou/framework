package rest

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

type ServiceRest struct {
	*worker.WorkerBase

	server *http.Server

	routes map[string]map[string]func(c *gin.Context, wc *worker.WorkerContexts)
}

func NewServiceRest(config *worker.WorkerConfig) *ServiceRest {
	routes := make(map[string]map[string]func(c *gin.Context, wc *worker.WorkerContexts))
	return &ServiceRest{WorkerBase: worker.NewWorkerBase(config), routes: routes}
}

func (w *ServiceRest) SetRoute(method string, path string, handler func(c *gin.Context, wc *worker.WorkerContexts)) {
	_, ok := w.routes[strings.ToUpper(method)]

	if !ok {
		w.routes[strings.ToUpper(method)] = (make(map[string]func(c *gin.Context, wc *worker.WorkerContexts)))
	}

	w.routes[strings.ToUpper(method)][path] = handler
}

func (w *ServiceRest) Setup() {
	w.Logger.Infof("Setting up REST Service")

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	router := gin.Default()

	router.Use(func(c *gin.Context) {
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

	if paths, ok := w.routes["GET"]; ok {
		for path, handler := range paths {
			router.GET(path, func(c *gin.Context) {
				handler(c, w.Contexts)
			})
		}
	}

	if paths, ok := w.routes["POST"]; ok {
		for path, handler := range paths {
			router.POST(path, func(c *gin.Context) {
				handler(c, w.Contexts)
			})
		}
	}

	if paths, ok := w.routes["PUT"]; ok {
		for path, handler := range paths {
			router.PUT(path, func(c *gin.Context) {
				handler(c, w.Contexts)
			})
		}
	}

	if paths, ok := w.routes["PATCH"]; ok {
		for path, handler := range paths {
			router.PATCH(path, func(c *gin.Context) {
				handler(c, w.Contexts)
			})
		}
	}

	if paths, ok := w.routes["DELETE"]; ok {
		for path, handler := range paths {
			router.DELETE(path, func(c *gin.Context) {
				handler(c, w.Contexts)
			})
		}
	}

	if paths, ok := w.routes["OPTIONS"]; ok {
		for path, handler := range paths {
			router.OPTIONS(path, func(c *gin.Context) {
				handler(c, w.Contexts)
			})
		}
	}

	w.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", w.Config.ListenHost, w.Config.Port),
		Handler: router,
	}
}

func (w *ServiceRest) Run() {
	w.Logger.Infof("Running REST Service")

	if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		w.Logger.Fatalf("listen %s\n", err)
	}
}

func (w *ServiceRest) Stop() {
	w.Logger.Infof("stop signal received! Graceful shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.server.Shutdown(ctx); err != nil {
		w.Logger.Fatal("Server forced to shutdown: ", err)
	}
}
