package worker

import (
	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"
)

type WorkerInterface interface {
	GetName() string
	SetName(string)
	SetAdapter(adapter.AdapterInterface)
	SetupAdapters() error
	CloseAdapters() error
	Setup()
	Run()
	Stop()
}

type BaseWorker struct {
	name     string
	Logger   *logrus.Entry
	Adapters *WorkerAdapters
}

func NewBaseWorker(name string) *BaseWorker {
	if name == "" {
		name = "default"
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &BaseWorker{
		name:     name,
		Logger:   logger.WithField("worker", name),
		Adapters: NewWorkerAdapters(),
	}
}

func (w *BaseWorker) GetName() string {
	return w.name
}
func (w *BaseWorker) SetName(name string) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	w.Logger = logger.WithField("worker", name)

	w.name = name
}

func (w *BaseWorker) SetAdapter(adap adapter.AdapterInterface) {
	w.Adapters.SetAdapter(adap)
}

func (w *BaseWorker) SetupAdapters() error {
	return w.Adapters.SetupAdapters()
}
func (w *BaseWorker) CloseAdapters() error {
	return w.Adapters.CloseAdapters()
}
