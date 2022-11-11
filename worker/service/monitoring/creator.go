package monitoring

import (
	"errors"
	"fmt"

	"github.com/radianteam/framework"
	"github.com/radianteam/framework/adapter/util/config"
)

func MonitoringMicroserviceCreate(name string, configAdapter *config.ConfigAdapter) (*framework.RadianMicroservice, error) {
	if configAdapter == nil {
		return nil, errors.New("configuration adapter is not provided")
	}

	monitoringConfig := &MonitoringServiceConfig{}
	err := configAdapter.UnmarshalPath([]string{}, monitoringConfig, false)

	if err != nil {
		return nil, fmt.Errorf("configuration loading error for the service with name %s: %v", name, err)
	}

	// microservice instance
	microservice := framework.NewRadianMicroservice(name)
	workerPrometheus := NewMonitoringServiceWorker("monitoring_service", monitoringConfig)

	err = microservice.AddWorker(workerPrometheus)

	if err != nil {
		return nil, fmt.Errorf("monitoring add worker error for the service with name %s: %v", name, err)
	}

	return microservice, nil
}
