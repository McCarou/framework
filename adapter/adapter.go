package adapter

import "github.com/sirupsen/logrus"

// Adapter structure contains an adapter name. All new adapters
// must inherit BaseAdapter and implement only Setup() and Close()
// functions from AdapterInterface.
type BaseAdapter struct {
	name   string
	Logger *logrus.Entry
}

// Interface implements basic adapter functions. All new adapters
// must inherit BaseAdapter and implement only Setup() and Close()
// functions.
type AdapterInterface interface {
	Setup() error
	Close() error
	GetName() string
	SetLogger(logger *logrus.Entry)
}

// Function allocates BaseAdapter structure with the name.
func NewBaseAdapter(name string) *BaseAdapter {
	if name == "" {
		name = "default"
	}

	return &BaseAdapter{
		name: name,
	}
}

// Function returns the name of an adapter.
func (a *BaseAdapter) GetName() string {
	return a.name
}

// Function sets the logger for internal using in the adapter.
func (a *BaseAdapter) SetLogger(logger *logrus.Entry) {
	a.Logger = logger
}
