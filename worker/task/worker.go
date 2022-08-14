package task

import (
	"context"
	"time"

	"github.com/procyon-projects/chrono"
	"github.com/radianteam/framework/worker"
)

type TaskType int8

const (
	TaskTypeFixedDelay TaskType = iota
	TaskTypeDelay
	TaskTypeCron
)

type HandleFuncTaskScheduler func(ctx context.Context, wc *worker.WorkerAdapters) error

type Task struct {
	Type TaskType

	Delay   time.Duration
	CronStr string

	Executor *chrono.Task
	Handler  HandleFuncTaskScheduler
}

const WorkerTaskSchedule string = "_worker_task_schedule"

type TaskSchedule struct {
	*worker.BaseWorker

	scheduler chrono.TaskScheduler

	waitChan chan bool

	tasks []Task
}

func NewTaskSchedule(name string) *TaskSchedule {
	return &TaskSchedule{BaseWorker: worker.NewBaseWorker(name)}
}

func (w *TaskSchedule) AddDelayTask(delay time.Duration, handler func(ctx context.Context, wc *worker.WorkerAdapters) error) {
	w.tasks = append(w.tasks, Task{Type: TaskTypeDelay, Delay: delay, Handler: handler})
}

func (w *TaskSchedule) AddFixedDelayTask(delay time.Duration, handler func(ctx context.Context, wc *worker.WorkerAdapters) error) {
	w.tasks = append(w.tasks, Task{Type: TaskTypeFixedDelay, Delay: delay, Handler: handler})
}

func (w *TaskSchedule) AddCronTask(cronStr string, handler func(ctx context.Context, wc *worker.WorkerAdapters) error) {
	w.tasks = append(w.tasks, Task{Type: TaskTypeCron, CronStr: cronStr, Handler: handler})
}

func (w *TaskSchedule) Setup() {
	w.Logger.Info("Setting up Task Scheduler")

	w.scheduler = chrono.NewDefaultTaskScheduler()
}

func (w *TaskSchedule) Run() {
	w.Logger.Info("Running Task scheduler")

	for _, task := range w.tasks {
		taskScope := task

		handler := func(ctx context.Context) {
			err := taskScope.Handler(ctx, w.Adapters)

			if err != nil {
				w.Logger.Errorf("Task has been completed with error: %v", err)
			}
		}

		if task.Type == TaskTypeFixedDelay {
			w.scheduler.ScheduleWithFixedDelay(handler, taskScope.Delay)
		} else if task.Type == TaskTypeDelay {
			w.scheduler.ScheduleAtFixedRate(handler, taskScope.Delay)
		} else if task.Type == TaskTypeCron {
			w.scheduler.ScheduleWithCron(handler, taskScope.CronStr)
		}
	}

	w.waitChan = make(chan bool)

	<-w.waitChan

	w.Logger.Info("Stopping Task Scheduler")

	wait := w.scheduler.Shutdown()

	<-wait
}

func (w *TaskSchedule) Stop() {
	w.Logger.Info("stop signal received! Graceful shutting down")

	close(w.waitChan)
}
