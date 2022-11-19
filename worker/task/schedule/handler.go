package schedule

import (
	"context"

	"github.com/radianteam/framework/worker"
)

type TaskScheduleHandlerInterface interface {
	worker.BaseHandlerInterface

	SetContext(ctx context.Context)
}

type TaskScheduleHandler struct {
	worker.BaseHandler

	ctx context.Context
}

func (h *TaskScheduleHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}
