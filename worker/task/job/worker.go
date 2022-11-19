package job

import (
	"context"

	"github.com/radianteam/framework/worker"
)

// Structure contains a task handler.
type TaskJob struct {
	*worker.BaseWorker

	Handler TaskJobHandlerInterface
}

// Function creates a new job.
func NewTaskJob(name string, handler TaskJobHandlerInterface) *TaskJob {
	return &TaskJob{BaseWorker: worker.NewBaseWorker(name), Handler: handler}
}

// Internal function. Main loop used in framework loop as a separated thread
func (w *TaskJob) Run(ctx context.Context) (err error) {
	w.Logger.Infof("Running Job: %s", w.GetName())

	w.Handler.SetContext(ctx)
	w.Handler.SetAdapters(w.Adapters)
	w.Handler.SetLogger(w.Logger.WithField("job", w.GetName()))

	err = w.Handler.Handle()

	if err != nil {
		w.Logger.Infof("Job %s has been completed with error: %v", w.GetName(), err)
		return
	}

	w.Logger.Infof("Job %s has been successfully completed", w.GetName())

	return
}
