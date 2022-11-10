package framework

// TODO: healthchecks
// TODO: worker rabbitmq reconnect
// TODO: adapter email
// TODO: tests
// TODO: make permanent workers
// TODO: remove fatal behaviour and return error everywhere

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
)

type MicroserviceMap map[string]*RadianMicroservice

// Microservice orchestrator's structure that holds microservices.
type RadianServiceManager struct {
	microservices     MicroserviceMap
	microserviceNames []string

	logger *logrus.Entry
}

// Function allocates structure with global JSON logger and an empty
// (but not nil!) worker list.
func NewRadianServiceManager() *RadianServiceManager {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	mmap := make(MicroserviceMap)

	return &RadianServiceManager{
		microservices: mmap,
		logger:        logger.WithField("manager", "framework"),
	}
}

// AddMicroservice registers a microservice by name from
// Microservice.GetName(). If a microservice with the same
// name is already registred an error will be thrown.
func (rsm *RadianServiceManager) AddMicroservice(ms *RadianMicroservice) error {
	if _, ok := rsm.microservices[ms.GetName()]; ok {
		return fmt.Errorf("microservice with name %s has been already registered", ms.GetName())
	}

	if rsm.microservices == nil {
		rsm.microservices = make(MicroserviceMap)
	}

	rsm.microservices[ms.GetName()] = ms

	rsm.microserviceNames = append(rsm.microserviceNames, ms.GetName())

	return nil
}

// Main framework loop. Runs all microservices. The loop setups
// microservices, captures the thread and wait for SIGINT or SIGTERM
// signals. After termination releases the thread.
func (rsm *RadianServiceManager) RunAll() {
	rsm.Run(rsm.microserviceNames)
}

// Main framework loop. The loop runs microservices in different
// goroutines, captures the thread and wait for SIGINT or SIGTERM
// signals. After termination releases the thread.
func (rsm *RadianServiceManager) Run(_microservices []string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	rsm.logger.Info("running")

	// check microservice names
	for _, serviceName := range _microservices {
		if _, ok := rsm.microservices[serviceName]; !ok {
			rsm.logger.Fatalf("worker with name %s is not found", serviceName)
		}
	}

	// run workers
	wg := sync.WaitGroup{}

	for _, microserviceName := range _microservices {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			rsm.logger.Infof("microservice %s: running", name)
			rsm.microservices[name].RunAll() // TODO: return and check error

			rsm.logger.Infof("microservice %s: stopped", name)
		}(microserviceName)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	rsm.logger.Info("stopping workers")

	wg.Wait()

	rsm.logger.Info("stopped")
}
