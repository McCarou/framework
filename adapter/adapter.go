package adapter

// Adapter structure contains an adapter name. All new adapters
// must inherit BaseAdapter and implement only Setup() and Close()
// functions from AdapterInterface.
type BaseAdapter struct {
	name string
}

// Interface implements basic adapter functions. All new adapters
// must inherit BaseAdapter and implement only Setup() and Close()
// functions.
type AdapterInterface interface {
	Setup() error
	Close() error
	GetName() string
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
