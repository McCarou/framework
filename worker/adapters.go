package worker

import (
	"errors"

	"github.com/radianteam/framework/adapter"
)

type AdaptersMap map[string]adapter.AdapterInterface

type WorkerAdapters struct {
	adapters AdaptersMap
}

func NewWorkerAdapters() *WorkerAdapters {
	adapters := make(AdaptersMap)
	return &WorkerAdapters{adapters: adapters}
}

func (w *WorkerAdapters) SetAdapter(adapter adapter.AdapterInterface) {
	w.adapters[adapter.GetName()] = adapter
}

func (w *WorkerAdapters) SetupAdapters() (err error) {
	for _, element := range w.adapters {
		err = element.Setup()
		if err != nil {
			return
		}
	}

	return
}

func (w *WorkerAdapters) CloseAdapters() (err error) {
	for _, adap := range w.adapters {

		if err = adap.Close(); err != nil {
			return
		}
	}

	return
}

func (w *WorkerAdapters) Get(name string) (adapter.AdapterInterface, error) {
	if val, ok := w.adapters[name]; ok {
		return val, nil
	}

	return nil, errors.New("adapter is not found")
}
