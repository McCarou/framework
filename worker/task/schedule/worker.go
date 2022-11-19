package schedule

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

// Structure contains information about chrono task
// (not pretasks and posttasks!).
type Task struct {
	Type TaskType

	Delay   time.Duration
	CronStr string

	Executor *chrono.Task
	Handler  TaskScheduleHandlerInterface
}

// Structure contains a task list and chrono scheduler
// to control the tasks.
type TaskSchedule struct {
	*worker.BaseWorker

	scheduler chrono.TaskScheduler

	waitChan chan bool

	tasks []Task
}

// Function creates a new chrono task worker.
func NewTaskSchedule(name string) *TaskSchedule {
	return &TaskSchedule{BaseWorker: worker.NewBaseWorker(name)}
}

// Function adds a delay task. The task will be
// executed after fixed duration AFTER completing previous execution.
func (w *TaskSchedule) AddDelayTask(delay time.Duration, handler TaskScheduleHandlerInterface) {
	w.tasks = append(w.tasks, Task{Type: TaskTypeDelay, Delay: delay, Handler: handler})
}

// Function adds a fixed delay task. The task will be
// executed after fixed duration FROM previous execution start.
func (w *TaskSchedule) AddFixedDelayTask(delay time.Duration, handler TaskScheduleHandlerInterface) {
	w.tasks = append(w.tasks, Task{Type: TaskTypeFixedDelay, Delay: delay, Handler: handler})
}

// Function adds a fixed delay task. The task will be
// executed by a cron schedule rule
func (w *TaskSchedule) AddCronTask(cronStr string, handler TaskScheduleHandlerInterface) {
	w.tasks = append(w.tasks, Task{Type: TaskTypeCron, CronStr: cronStr, Handler: handler})
}

// Internal function to execute during the framework starting
func (w *TaskSchedule) Setup() {
	w.Logger.Info("Setting up Task Scheduler")

	w.scheduler = chrono.NewDefaultTaskScheduler()
}

// Internal function. Main loop used in framework loop as a separated thread
func (w *TaskSchedule) Run() {
	w.Logger.Info("Running Task scheduler")

	for _, task := range w.tasks {
		taskScope := task

		handler := func(ctx context.Context) {
			taskScope.Handler.SetAdapters(w.Adapters)
			taskScope.Handler.SetContext(ctx)

			err := taskScope.Handler.Handle()

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

// Internal function to execute during the framework stopping
func (w *TaskSchedule) Stop() {
	w.Logger.Info("stop signal received! Graceful shutting down")

	close(w.waitChan)
}
