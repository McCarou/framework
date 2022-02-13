package worker

import (
	"errors"

	"github.com/radianteam/framework/context"
)

type WorkerContexts struct {
	contexts map[string]context.ContextInterface
}

func NewWorkerContexts() *WorkerContexts {
	ctx_list := make(map[string]context.ContextInterface)
	return &WorkerContexts{contexts: ctx_list}
}

func (w *WorkerContexts) AddContext(name string, ctx context.ContextInterface) {
	w.contexts[name] = ctx
}

func (w *WorkerContexts) SetupContexts() error {
	for _, element := range w.contexts {
		err := element.Setup()

		if err != nil {
			return err
		}
	}

	return nil
}

func (w *WorkerContexts) CloseContexts() error {
	for _, element := range w.contexts {
		err := element.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func (w *WorkerContexts) Get(name string) (interface{}, error) {
	if val, ok := w.contexts[name]; ok {
		return val, nil
	}

	return nil, errors.New("context is not found")
}
