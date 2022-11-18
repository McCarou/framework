package worker

import (
	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"
)

// Interface implements basic worker functions. All new workers
// must inherit BaseWorker structure and implement only Setup(),
// Run() and Stop() functions.
type WorkerInterface interface {
	GetName() string
	SetName(string)
	SetMicroserviceName(string)
	SetAdapter(adapter.AdapterInterface)
	SetupAdapters() error
	CloseAdapters() error
	Setup()
	Run()
	Stop()
	SetMonitoring(enabled bool)
	IsMonitoringEnable() bool
}

// Worker structure contains an adapter list and implements
// functions to control adapters. All new workers must inherit
// BaseWorker and implement only Setup(), Run() and Stop()
// functions from WorkerInterface.
type BaseWorker struct {
	name              string
	monitoringEnabled bool
	Logger            *logrus.Entry
	Adapters          *WorkerAdapters
}

// Function allocates BaseWorker structure with JSON logger
// and an empty (but not nil!) adapter list.
func NewBaseWorker(name string) *BaseWorker {
	if name == "" {
		name = "default"
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &BaseWorker{
		name:              name,
		Logger:            logger.WithField("worker", name),
		Adapters:          NewWorkerAdapters(),
		monitoringEnabled: false,
	}
}

// Get current worker name.
func (w *BaseWorker) GetName() string {
	return w.name
}

// Set current worker name.
func (w *BaseWorker) SetName(name string) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	w.Logger = logger.WithField("worker", name)

	w.name = name
}

// Set current worker name.
func (w *BaseWorker) SetMicroserviceName(name string) {
	w.Logger = w.Logger.WithField("microservice", name)
}

// Function appends an adapter to the worker's adapter list.
// If the adapter with the same name is already registred the
// first one will be overwritten by the new one.
func (w *BaseWorker) SetAdapter(adap adapter.AdapterInterface) {
	w.Adapters.SetAdapter(adap)
}

// Function setups all adapters and is used in the main
// framework loop.
func (w *BaseWorker) SetupAdapters() error {
	return w.Adapters.SetupAdapters()
}

// Function clears all adapters and is used in the main
// framework loop.
func (w *BaseWorker) CloseAdapters() error {
	return w.Adapters.CloseAdapters()
}

// Function enables or disables pushing metrics
// to the prometheus.
func (w *BaseWorker) SetMonitoring(enabled bool) {
	w.monitoringEnabled = enabled
}

// Function returns monitoring status.
func (w *BaseWorker) IsMonitoringEnable() bool {
	return w.monitoringEnabled
}
