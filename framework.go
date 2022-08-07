package framework

// TODO: healthchecks
// TODO: worker rabbitmq reconnect
// TODO: context email
// TODO: tests
// TODO: make workers for periodic, permanent, pretask and posttask

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/radianteam/framework/worker"
)

type WorkersMap map[string]worker.WorkerInterface

type RadianFramework struct {
	workers WorkersMap
	logger  *logrus.Entry
}

func NewRadianFramework() *RadianFramework {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	w := make(WorkersMap)

	return &RadianFramework{
		workers: w,
		logger:  logger.WithField("worker", "framework"),
	}
}

func (r *RadianFramework) AddWorker(w worker.WorkerInterface) {
	if r.workers == nil {
		r.workers = make(WorkersMap)
	}

	r.workers[w.GetName()] = w
}

func (r *RadianFramework) Run(services []string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	r.logger.Info("running")

	// check service names
	for _, serviceName := range services {
		if _, ok := r.workers[serviceName]; !ok {
			r.logger.Fatalf("worker with name %s is not found", serviceName)
		}
	}

	// run services
	wg := sync.WaitGroup{}

	for _, serviceName := range services {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			r.logger.Infof("worker %s: setting up adapters", name)
			err := r.workers[name].SetupAdapters()

			if err != nil {
				r.logger.Fatalf("worker %s: adapter init error %v", name, err)
			}

			r.logger.Infof("worker %s: setting up worker", name)
			r.workers[name].Setup()

			r.logger.Infof("worker %s: running", name)
			r.workers[name].Run()

			r.logger.Infof("worker %s: stopping", name)
			r.logger.Infof("worker %s: deleting adapters", name)
			err = r.workers[name].CloseAdapters()

			if err != nil {
				r.logger.Fatalf("worker %s: adapter close error %v", name, err)
			}

			r.logger.Infof("worker %s: stopped", name)
		}(serviceName)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	r.logger.Info("stopping workers")

	for _, serviceName := range services {
		r.workers[serviceName].Stop()
	}

	wg.Wait()

	r.logger.Info("stopped")
}
