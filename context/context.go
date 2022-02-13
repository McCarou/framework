package context

type ContextInterface interface {
	Setup() error
	Close() error
	Get() interface{}
}
