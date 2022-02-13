package framework

//TODO: healthchecks

//TODO: worker rabbitmq reconnect

//TODO: context email

//TODO: tests

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/radianteam/framework/worker"
)

type RadianFramework struct {
	workers map[string]worker.WorkerInterface
}

func NewRadianFramework() *RadianFramework {
	m := make(map[string]worker.WorkerInterface)
	return &RadianFramework{workers: m}
}

func (r *RadianFramework) AddWorker(name string, worker worker.WorkerInterface) {
	worker.SetName(name)
	r.workers[name] = worker
}

func (r *RadianFramework) Run(services []string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	log_struct := logrus.New()
	log_struct.SetFormatter(&logrus.JSONFormatter{})
	logger := log_struct.WithField("worker", "framework")

	logger.Info("running")

	// check service names
	for _, item := range services {
		if _, ok := r.workers[item]; !ok {
			logger.Fatal("worker with name %s is not found", item)
		}
	}

	// run services
	wg := sync.WaitGroup{}

	for _, item := range services {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			logger.Infof("worker %s: setting up contexts", name)
			err := r.workers[name].SetupContexts()

			if err != nil {
				logger.Fatalf("worker %s: context init error %v", name, err)
			}

			logger.Infof("worker %s: setting up worker", name)
			r.workers[name].Setup()

			logger.Infof("worker %s: running", name)
			r.workers[name].Run()

			logger.Infof("worker %s: stopping", name)
			logger.Infof("worker %s: deleting contexts", name)
			err = r.workers[name].CloseContexts()

			if err != nil {
				logger.Fatalf("worker %s: context close error %v", name, err)
			}

			logger.Infof("worker %s: stopped", name)
		}(item)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	logger.Info("stopping workers")

	for _, item := range services {
		r.workers[item].Stop()
	}

	wg.Wait()

	logger.Info("stopped")
}
