package job

import (
	"context"

	"github.com/radianteam/framework/worker"
	"github.com/sirupsen/logrus"
)

// Function template to handle tasks.
type TaskJobHandleFunc func(ctx context.Context, wc *worker.WorkerAdapters) error

// Structure contains a task handler.
type TaskJob struct {
	*worker.BaseWorker

	Handler TaskJobHandleFunc
}

// Function creates a new job.
func NewTaskJob(name string, handler TaskJobHandleFunc) *TaskJob {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &TaskJob{BaseWorker: worker.NewBaseWorker(name), Handler: handler}
}

// Internal function. Main loop used in framework loop as a separated thread
func (w *TaskJob) Run(ctx context.Context) (err error) {
	w.Logger.Infof("Running Job: %s", w.GetName())

	err = w.Handler(ctx, w.Adapters)

	if err != nil {
		w.Logger.Infof("Job %s has been completed with error: %v", err)
		return
	}

	w.Logger.Infof("Job %s has been successfully completed", w.GetName())

	return
}