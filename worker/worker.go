package worker

import (
	"github.com/radianteam/framework/context"
	"github.com/sirupsen/logrus"
)

type WorkerInterface interface {
	SetName(string)
	SetupContexts() error
	CloseContexts() error
	Setup()
	Run()
	Stop()
}

type WorkerConfig struct {
	ListenHost string
	Host       string
	Port       int16
	Login      string
	Password   string
	Threads    int8
}

type WorkerBase struct {
	Config *WorkerConfig
	Name   string

	Logger *logrus.Entry

	Contexts *WorkerContexts
}

func NewWorkerBase(config *WorkerConfig) *WorkerBase {
	return &WorkerBase{Config: config, Contexts: NewWorkerContexts()}
}

func (w *WorkerBase) SetConfig(config *WorkerConfig) {
	w.Config = config
}

func (w *WorkerBase) SetName(name string) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	w.Logger = logger.WithField("worker", name)

	w.Name = name
}

func (w *WorkerBase) AddContext(name string, ctx context.ContextInterface) {
	w.Contexts.AddContext(name, ctx)
}

func (w *WorkerBase) SetupContexts() error {
	return w.Contexts.SetupContexts()
}

func (w *WorkerBase) CloseContexts() error {
	return w.Contexts.CloseContexts()
}
