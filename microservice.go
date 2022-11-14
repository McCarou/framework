package framework

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/radianteam/framework/worker"
	"github.com/radianteam/framework/worker/task/job"
)

type WorkersMap map[string]worker.WorkerInterface
type JobsMap map[string]*job.TaskJob

// Main Microservice structure that holds worker list and global logger.
type RadianMicroservice struct {
	name string

	preJobs      JobsMap
	preJobNames  []string
	postJobs     JobsMap
	postJobNames []string

	workers     WorkersMap
	workerNames []string

	logger *logrus.Entry
}

// Function allocates structure with global JSON logger and an empty
// (but not nil!) worker list.
func NewRadianMicroservice(name string) *RadianMicroservice {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	w := make(WorkersMap)

	return &RadianMicroservice{
		workers: w,
		name:    name,
		logger:  logger.WithField("worker", name),
	}
}

// AddWorker registers a worker by name from Worker.GetName().
// If a worker with the same name is already registred an
// error will be thrown
func (r *RadianMicroservice) AddWorker(w worker.WorkerInterface) error {
	if _, ok := r.workers[w.GetName()]; ok {
		return fmt.Errorf("worker with name %s has been already registered", w.GetName())
	}

	if r.workers == nil {
		r.workers = make(WorkersMap)
	}

	r.workers[w.GetName()] = w

	r.workerNames = append(r.workerNames, w.GetName())

	return nil
}

// AddPreJob registers a list of functions that will be executed
// before starting the main loop. Use it to get tokens, make
// migrations, etc.
func (r *RadianMicroservice) AddPreJob(t *job.TaskJob) error {
	if _, ok := r.preJobs[t.GetName()]; ok {
		return fmt.Errorf("prejob with name %s has been already registered", t.GetName())
	}

	if r.preJobs == nil {
		r.preJobs = make(JobsMap)
	}

	r.preJobs[t.GetName()] = t

	r.preJobNames = append(r.preJobNames, t.GetName())

	return nil
}

// AddPostJob registers a list of functions that will be executed
// after finishing the main loop. Use it to invalidate tokens, make
// termination signals, etc.
func (r *RadianMicroservice) AddPostJob(t *job.TaskJob) error {
	if _, ok := r.postJobs[t.GetName()]; ok {
		return fmt.Errorf("postjob with name %s has been already registered", t.GetName())
	}

	if r.postJobs == nil {
		r.postJobs = make(JobsMap)
	}

	r.postJobs[t.GetName()] = t

	r.postJobNames = append(r.postJobNames, t.GetName())

	return nil
}

// Returns the name of a microservice
func (r *RadianMicroservice) GetName() string {
	return r.name
}

// Main microservice loop. Left for backward compatibility. Doesn't
// run jobs. Use RunWithJobs() instead of Run(). The loop setups
// adapters, captures the thread and wait for SIGINT or SIGTERM
// signals. After termination runs postjobs and releases the thread.
func (r *RadianMicroservice) Run(_workers []string) {
	r.RunWithJobs([]string{}, _workers, []string{})
}

// Main microservice loop. Runs all prejobs, workers and postjobs. The loop setups
// adapters, captures the thread and wait for SIGINT or SIGTERM
// signals. After termination runs postjobs and releases the thread.
func (r *RadianMicroservice) RunAll() {
	r.RunWithJobs(r.preJobNames, r.workerNames, r.postJobNames)
}

// Main microservice loop. Use this instead of Run(). The loop
// setups adapters, runs prejobs, captures the thread and wait
// for SIGINT or SIGTERM signals. After termination runs
// postjobs and releases the thread.
func (r *RadianMicroservice) RunWithJobs(_preJobs []string, _workers []string, _postJobs []string) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	r.logger.Info("running")

	// check prejob names
	for _, jobName := range _preJobs {
		if _, ok := r.preJobs[jobName]; !ok {
			avails := []string{}
			for k := range r.preJobs {
				avails = append(avails, k)
			}

			r.logger.Fatalf("prejob with name %s is not found. Available names: %s", jobName, strings.Join(avails, ", "))
		}
	}

	// check service names
	for _, serviceName := range _workers {
		if _, ok := r.workers[serviceName]; !ok {
			avails := []string{}
			for k := range r.workers {
				avails = append(avails, k)
			}

			r.logger.Fatalf("worker with name %s is not found. Available names: %s", serviceName, strings.Join(avails, ", "))
		}
	}

	// check postjob names
	for _, jobName := range _postJobs {
		if _, ok := r.postJobs[jobName]; !ok {
			avails := []string{}
			for k := range r.postJobs {
				avails = append(avails, k)
			}

			r.logger.Fatalf("postjob with name %s is not found. Available names: %s", jobName, strings.Join(avails, ", "))
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
