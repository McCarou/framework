package job

import (
	"context"

	"github.com/radianteam/framework/worker"
)

type TaskJobHandlerInterface interface {
	worker.BaseHandlerInterface

	SetContext(ctx context.Context)
}

type TaskJobHandler struct {
	worker.BaseHandler

	ctx context.Context
}

func (h *TaskJobHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}
