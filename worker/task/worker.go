package task

import (
	"context"
	"time"

	"github.com/procyon-projects/chrono"
	"github.com/radianteam/framework/worker"
)

type TaskType int8

const (
	TASK_TYPE_FIXED_DELAY TaskType = iota
	TASK_TYPE_DELAY
	TASK_TYPE_CRON
)

type Task struct {
	Type TaskType

	Delay   time.Duration
	CronStr string

	Executor *chrono.Task
	Handler  func(ctx context.Context, wc *worker.WorkerContexts) error
}

type TaskSchedule struct {
	*worker.WorkerBase

	scheduler chrono.TaskScheduler

	wait_chan chan bool

	tasks []Task
}

func NewTaskSchedule() *TaskSchedule {
	return &TaskSchedule{WorkerBase: worker.NewWorkerBase(nil)}
}

func (w *TaskSchedule) AddDelayTask(delay time.Duration, handler func(ctx context.Context, wc *worker.WorkerContexts) error) {
	w.tasks = append(w.tasks, Task{Type: TASK_TYPE_DELAY, Delay: delay, Handler: handler})
}

func (w *TaskSchedule) AddFixedDelayTask(delay time.Duration, handler func(ctx context.Context, wc *worker.WorkerContexts) error) {
	w.tasks = append(w.tasks, Task{Type: TASK_TYPE_FIXED_DELAY, Delay: delay, Handler: handler})
}

func (w *TaskSchedule) AddCronTask(cron_str string, handler func(ctx context.Context, wc *worker.WorkerContexts) error) {
	w.tasks = append(w.tasks, Task{Type: TASK_TYPE_CRON, CronStr: cron_str, Handler: handler})
}

func (w *TaskSchedule) Setup() {
	w.Logger.Info("Setting up Task Scheduler")

	w.scheduler = chrono.NewDefaultTaskScheduler()
}

func (w *TaskSchedule) Run() {
	w.Logger.Info("Running Task scheduler")

	for _, task := range w.tasks {
		task_scope := task

		handler := func(ctx context.Context) {
			err := task_scope.Handler(ctx, w.Contexts)

			if err != nil {
				w.Logger.Errorf("Task has been completed with error: %v", err)
			}
		}

		if task.Type == TASK_TYPE_FIXED_DELAY {
			w.scheduler.ScheduleWithFixedDelay(handler, task_scope.Delay)
		} else if task.Type == TASK_TYPE_DELAY {
			w.scheduler.ScheduleAtFixedRate(handler, task_scope.Delay)
		} else if task.Type == TASK_TYPE_CRON {
			w.scheduler.ScheduleWithCron(handler, task_scope.CronStr)
		}
	}

	w.wait_chan = make(chan bool)

	<-w.wait_chan

	w.Logger.Info("Stopping Task Scheduler")

	wait := w.scheduler.Shutdown()

	<-wait
}

func (w *TaskSchedule) Stop() {
	w.Logger.Info("stop signal received! Graceful shutting down")

	close(w.wait_chan)
}
