package adapter

type BaseAdapter struct {
	name string
}

type AdapterInterface interface {
	Setup() error
	Close() error
	GetName() string
}

func NewBaseAdapter(name string) *BaseAdapter {
	if name == "" {
		name = "default"
	}

	return &BaseAdapter{
		name: name,
	}
}

func (a *BaseAdapter) GetName() string {
	return a.name
}
