package worker

import (
	"errors"

	"github.com/radianteam/framework/adapter"
)

type AdaptersMap map[string]adapter.AdapterInterface

// Structure holds adapter map.
type WorkerAdapters struct {
	adapters AdaptersMap
}

// Function allocates WorkerAdapters structure with an empty
// (but not nil!) adapter list.
func NewWorkerAdapters() *WorkerAdapters {
	adapters := make(AdaptersMap)
	return &WorkerAdapters{adapters: adapters}
}

// Function appends an adapter to the worker's adapter list.
// If the adapter with the same name is already registred the
// first one will be overwritten by the new one.
func (w *WorkerAdapters) SetAdapter(adapter adapter.AdapterInterface) {
	w.adapters[adapter.GetName()] = adapter
}

// Function setups all adapters and is used in the main
// framework loop.
func (w *WorkerAdapters) SetupAdapters() (err error) {
	for _, element := range w.adapters {
		err = element.Setup()
		if err != nil {
			return
		}
	}

	return
}

// Function clears all adapters and is used in the main
// framework loop.
func (w *WorkerAdapters) CloseAdapters() (err error) {
	for _, adap := range w.adapters {

		if err = adap.Close(); err != nil {
			return
		}
	}

	return
}

// Function receives an adapter interface that can be
// converted to a particular adapter structure.
func (w *WorkerAdapters) Get(name string) (adapter.AdapterInterface, error) {
	if val, ok := w.adapters[name]; ok {
		return val, nil
	}

	return nil, errors.New("adapter is not found")
}
