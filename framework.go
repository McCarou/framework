package framework

// TODO: healthchecks
// TODO: worker rabbitmq reconnect
// TODO: adapter email
// TODO: tests
// TODO: make permanent workers

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/radianteam/framework/worker"
	"github.com/radianteam/framework/worker/task/job"
)

type WorkersMap map[string]worker.WorkerInterface
type JobsMap map[string]*job.TaskJob

// Main framework structure that holds worker list and global logger.
type RadianFramework struct {
	preJobs  JobsMap
	postJobs JobsMap

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

// AddPreJob registers a list of functions that will be executed
// before starting the main loop. Use it to get tokens, make
// migrations, etc.
func (r *RadianFramework) AddPreJob(t *job.TaskJob) {
	if r.preJobs == nil {
		r.preJobs = make(JobsMap)
	}

	r.preJobs[t.GetName()] = t
}

// AddPostJob registers a list of functions that will be executed
// after finishing the main loop. Use it to invalidate tokens, make
// termination signals, etc.
func (r *RadianFramework) AddPostJob(t *job.TaskJob) {
	if r.postJobs == nil {
		r.postJobs = make(JobsMap)
	}

	r.postJobs[t.GetName()] = t
}

// Main framework loop. Left for backward compatibility. Doesn't
// run jobs. Use RunWithJobs() instead of Run(). The loop setups
// adapters, captures the thread and wait for SIGINT or SIGTERM
// signals. After termination runs postjobs and releases the thread.
func (r *RadianFramework) Run(_workers []string) {
	r.RunWithJobs([]string{}, _workers, []string{})
}

// Main framework loop. Runs all prejobs, workers and postjobs. The loop setups
// adapters, captures the thread and wait for SIGINT or SIGTERM
// signals. After termination runs postjobs and releases the thread.
func (r *RadianFramework) RunAll() {
	_preJobs := make([]string, 0, len(r.preJobs))
	for k := range r.preJobs {
		_preJobs = append(_preJobs, k)
	}

	_workers := make([]string, 0, len(r.workers))
	for k := range r.workers {
		_workers = append(_workers, k)
	}

	_postJobs := make([]string, 0, len(r.postJobs))
	for k := range r.postJobs {
		_postJobs = append(_postJobs, k)
	}

	r.RunWithJobs(_preJobs, _workers, _postJobs)
}

// Main framework loop. Use this instead of Run(). The loop
// setups adapters, runs prejobs, captures the thread and wait
// for SIGINT or SIGTERM signals. After termination runs
// postjobs and releases the thread.
func (r *RadianFramework) RunWithJobs(_preJobs []string, _workers []string, _postJobs []string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	r.logger.Info("running")

	// check prejob names
	for _, jobName := range _preJobs {
		if _, ok := r.preJobs[jobName]; !ok {
			r.logger.Fatalf("prejob with name %s is not found", jobName)
		}
	}

	// check service names
	for _, serviceName := range _workers {
		if _, ok := r.workers[serviceName]; !ok {
			r.logger.Fatalf("worker with name %s is not found", serviceName)
		}
	}

	// check postjob names
	for _, jobName := range _postJobs {
		if _, ok := r.postJobs[jobName]; !ok {
			r.logger.Fatalf("postjob with name %s is not found", jobName)
		}
	}

	// TODO: run prejobs
	for _, jobName := range _preJobs {
		r.logger.Infof("prejob %s: setting up adapters", jobName)
		err := r.preJobs[jobName].SetupAdapters()

		if err != nil {
			r.logger.Fatalf("prejob %s: adapter init error %v", jobName, err)
		}

		r.logger.Infof("prejob %s: running", jobName)
		err = r.preJobs[jobName].Run(context.TODO())

		if err != nil {
			r.logger.Fatalf("prejob %s: job run error %v", jobName, err)
		}

		r.logger.Infof("prejob %s: stopping", jobName)
		r.logger.Infof("prejob %s: deleting adapters", jobName)
		err = r.preJobs[jobName].CloseAdapters()

		if err != nil {
			r.logger.Fatalf("prejob %s: adapter close error %v", jobName, err)
		}

		r.logger.Infof("prejob %s: completed", jobName)
	}

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
			r.workers[name].Setup() // TODO: return and check error

			r.logger.Infof("worker %s: running", name)
			r.workers[name].Run() // TODO: return and check error

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

	// TODO: run posttasks
	for _, jobName := range _postJobs {
		r.logger.Infof("postjob %s: setting up adapters", jobName)
		err := r.postJobs[jobName].SetupAdapters()

		if err != nil {
			r.logger.Fatalf("postjob %s: adapter init error %v", jobName, err)
		}

		r.logger.Infof("postjob %s: running", jobName)
		err = r.postJobs[jobName].Run(context.TODO())

		if err != nil {
			r.logger.Fatalf("postjob %s: job run error %v", jobName, err)
		}

		r.logger.Infof("postjob %s: stopping", jobName)
		r.logger.Infof("postjob %s: deleting adapters", jobName)
		err = r.postJobs[jobName].CloseAdapters()

		if err != nil {
			r.logger.Fatalf("postjob %s: adapter close error %v", jobName, err)
		}

		r.logger.Infof("postjob %s: completed", jobName)
	}

	r.logger.Info("stopped")
}
