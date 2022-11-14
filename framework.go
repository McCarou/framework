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
	"strings"
	"sync"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/radianteam/framework/adapter/util/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type MicroserviceMap map[string]*RadianMicroservice

type MicroserviceCreatorFunc func(name string, configAdapter *config.ConfigAdapter) (*RadianMicroservice, error)

// Microservice orchestrator's structure that holds microservices.
type RadianServiceManager struct {
	microservices     MicroserviceMap
	microserviceNames []string

	microserviceCreators map[string]MicroserviceCreatorFunc

	desiredServiceNames []string

	mainConfig *config.ConfigAdapter

	logger *logrus.Entry
}

// Function allocates structure with global JSON logger and an empty
// (but not nil!) worker list.
func NewRadianServiceManager() *RadianServiceManager {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &RadianServiceManager{
		microservices: make(MicroserviceMap),
		mainConfig:    config.NewConfigAdapter("Config"),
		logger:        logger.WithField("manager", "framework"),
	}
}

// AddMicroservice registers a microservice by name from
// Microservice.GetName(). If a microservice with the same
// name is already registred an error will be thrown.
func (rsm *RadianServiceManager) AddMicroservice(ms *RadianMicroservice) error {
	if slices.Contains(rsm.microserviceNames, ms.GetName()) {
		return fmt.Errorf("microservice with name %s has been already registered", ms.GetName())
	}

	if rsm.microservices == nil {
		rsm.microservices = make(MicroserviceMap)
	}

	rsm.microservices[ms.GetName()] = ms

	rsm.microserviceNames = append(rsm.microserviceNames, ms.GetName())

	return nil
}

func (rsm *RadianServiceManager) AddMicroserviceCreator(name string, creator MicroserviceCreatorFunc) error {
	if slices.Contains(rsm.microserviceNames, name) {
		return fmt.Errorf("microservice with name %s has been already registered", name)
	}

	if rsm.microserviceCreators == nil {
		rsm.microserviceCreators = make(map[string]MicroserviceCreatorFunc)
	}

	rsm.microserviceCreators[name] = creator

	rsm.microserviceNames = append(rsm.microserviceNames, name)

	return nil
}

func (rsm *RadianServiceManager) SetupFromCommandLine() (err error) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	var argOpts struct {
		Config string `short:"c" long:"config" description:"A configuration file name"`
		Mode   string `short:"m" long:"mode" description:"all, monolith, empty string or service names comma separated"`
	}

	_, err = flags.ParseArgs(&argOpts, os.Args)

	if err != nil {
		return
	}

	return rsm.setupInternal(argOpts.Mode, argOpts.Config)
}

func (rsm *RadianServiceManager) SetupFromEnv(prefix string) (err error) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	confAdapter := config.NewConfigAdapter("temp")

	err = confAdapter.LoadFromEnv(prefix)

	if err != nil {
		return
	}

	return rsm.setupInternal(confAdapter.GetStringOrDefault([]string{"Mode"}, ""), confAdapter.GetStringOrDefault([]string{"Config"}, ""))
}

func (rsm *RadianServiceManager) setupInternal(mode string, _config string) (err error) {
	names := strings.Split(mode, ",")

	if mode != "all" {
		if mode == "" || mode == "monolith" {
			mode = "all"
		} else if len(names) > 1 {
			rsm.desiredServiceNames = names
		} else {
			rsm.desiredServiceNames = []string{mode}
		}
	}

	if _config != "" {
		logrus.Infof("Loading configuration from file: %s", _config)

		err = rsm.mainConfig.LoadFromFileJson(_config)

		if err != nil {
			return
		}
	}

	return
}

// Main framework loop. Runs all microservices. The loop setups
// microservices, captures the thread and wait for SIGINT or SIGTERM
// signals. After termination releases the thread.
func (rsm *RadianServiceManager) RunAll() {
	rsm.Run(rsm.microserviceNames)
}

func (rsm *RadianServiceManager) RunDesired() {
	if len(rsm.desiredServiceNames) == 0 {
		rsm.RunAll()

		return
	}

	rsm.Run(rsm.desiredServiceNames)
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
			if _, ok := rsm.microserviceCreators[serviceName]; !ok {
				availWorkers := []string{}
				for k := range rsm.microservices {
					availWorkers = append(availWorkers, k)
				}
				for k := range rsm.microserviceCreators {
					availWorkers = append(availWorkers, k)
				}

				rsm.logger.Fatalf("worker with name %s is not found. Use \"all\" or names: %s", serviceName, strings.Join(availWorkers, ", "))
			}

			ms, err := rsm.microserviceCreators[serviceName](serviceName, rsm.mainConfig.GetAdapterOrNil([]string{serviceName}))

			if err != nil {
				rsm.logger.Fatalf("worker with name %s created with the error: %v", serviceName, err)
			}

			rsm.microservices[serviceName] = ms
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
