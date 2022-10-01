package framework

// TODO: healthchecks
// TODO: worker rabbitmq reconnect
// TODO: context email
// TODO: tests
// TODO: make workers for periodic, permanent, pretask and posttask

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/radianteam/framework/worker"
	"github.com/radianteam/framework/worker/task/job"
)

type WorkersMap map[string]worker.WorkerInterface

// Main framework structure that holds worker list and global logger.
type RadianFramework struct {
	preTasks  []*job.TaskJob
	postTasks []*job.TaskJob

	workers WorkersMap
	logger  *logrus.Entry
}

// Function allocates structure with global JSON logger and an empty
// (but not nil!) worker list.
func NewRadianFramework() *RadianFramework {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	w := make(WorkersMap)

	return &RadianFramework{
		workers: w,
		logger:  logger.WithField("worker", "framework"),
	}
}

// AddWorker registers a worker by name from Worker.GetName().
// If the worker with the same name is already registred the
// first one will be overwritten by the new one.
func (r *RadianFramework) AddWorker(w worker.WorkerInterface) {
	if r.workers == nil {
		r.workers = make(WorkersMap)
	}

	r.workers[w.GetName()] = w
}

// AddPreTask registers a list of functions that will be executed
// before starting the main loop. Use it to get tokens, make
// migrations, etc.
func (r *RadianFramework) AddPreTask(t *job.TaskJob) {
	r.preTasks = append(r.preTasks, t)
}

// AddPostTask registers a list of functions that will be executed
// after finishing the main loop. Use it to invalidate tokens, make
// termination signals, etc.
func (r *RadianFramework) AddPostTask(t *job.TaskJob) {
	r.postTasks = append(r.postTasks, t)
}

// Main framework loop. The loop setups adapters, runs pretasks,
// captures the, thread and wait for SIGINT or SIGTERM signals.
// After termination runs posttasks and releases the thread.
func (r *RadianFramework) Run(_preTasks []string, _workers []string, _postTasks []string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	r.logger.Info("running")

	// check pretask names
	for _, jobName := range _preTasks {
		if _, ok := r.workers[jobName]; !ok {
			r.logger.Fatalf("pretask with name %s is not found", jobName)
		}
	}

	// check service names
	for _, serviceName := range _workers {
		if _, ok := r.workers[serviceName]; !ok {
			r.logger.Fatalf("worker with name %s is not found", serviceName)
		}
	}

	// check posttask names
	for _, jobName := range _postTasks {
		if _, ok := r.workers[jobName]; !ok {
			r.logger.Fatalf("posttask with name %s is not found", jobName)
		}
	}

	// run pretasks

	// run workers
	wg := sync.WaitGroup{}

	for _, serviceName := range _workers {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			r.logger.Infof("worker %s: setting up adapters", name)
			err := r.workers[name].SetupAdapters()

			if err != nil {
				r.logger.Fatalf("worker %s: adapter init error %v", name, err)
			}

			r.logger.Infof("worker %s: setting up worker", name)
			r.workers[name].Setup()

			r.logger.Infof("worker %s: running", name)
			r.workers[name].Run()

			r.logger.Infof("worker %s: stopping", name)
			r.logger.Infof("worker %s: deleting adapters", name)
			err = r.workers[name].CloseAdapters()

			if err != nil {
				r.logger.Fatalf("worker %s: adapter close error %v", name, err)
			}

			r.logger.Infof("worker %s: stopped", name)
		}(serviceName)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	r.logger.Info("stopping workers")

	for _, serviceName := range _workers {
		r.workers[serviceName].Stop()
	}

	wg.Wait()

	r.logger.Info("stopped")
}
