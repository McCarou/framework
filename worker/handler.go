package worker

import (
	"errors"

	"github.com/sirupsen/logrus"
)

type BaseHandlerInterface interface {
	Handle() error
	SetLogger(l *logrus.Entry)
	SetAdapters(a *WorkerAdapters)
}

type BaseHandler struct {
	Logger   *logrus.Entry
	Adapters *WorkerAdapters
}

func (h *BaseHandler) Handle() error {
	return errors.New("handler is not implemented")
}

func (h *BaseHandler) SetLogger(l *logrus.Entry) {
	h.Logger = l
}

func (h *BaseHandler) SetAdapters(a *WorkerAdapters) {
	h.Adapters = a
}
