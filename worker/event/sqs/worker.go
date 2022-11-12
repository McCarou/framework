package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	sqs_adapter "github.com/radianteam/framework/adapter/event/sqs"
	"github.com/radianteam/framework/worker"
	"sync"
	"time"
)

const RetryConsumeTimeoutMs = 10000

type SqsWorkerRegFunc func(d *sqs.Message, wc *worker.WorkerAdapters) error

type SqsEventsWorker struct {
	*worker.BaseWorker

	config *sqs_adapter.SqsConfig

	mutex       sync.Mutex
	waitChan    chan struct{}
	stopPolling bool

	handlers map[string]SqsWorkerRegFunc
}

func NewSqsEventsWorker(name string, config *sqs_adapter.SqsConfig) *SqsEventsWorker {
	handlers := make(map[string]SqsWorkerRegFunc)

	return &SqsEventsWorker{BaseWorker: worker.NewBaseWorker(name), config: config, handlers: handlers}
}

func (w *SqsEventsWorker) SetEvent(queue string, handler SqsWorkerRegFunc) {
	w.handlers[queue] = handler
}

func (w *SqsEventsWorker) Setup() {
	w.Logger.Info("Setting up Sqs Events")
}

func (w *SqsEventsWorker) Run() {
	w.Logger.Info("Running Sqs Events Worker")

	wg := sync.WaitGroup{}

	adapter := sqs_adapter.NewSqsAdapter("sqs-consumer", w.config)
	err := adapter.Setup()
	if err != nil {
		w.Logger.Errorf("Failed to run '%s' worker during adapter configuration: %s", w.GetName(), err)

		return
	}

	for queueName, handler := range w.handlers {
		wg.Add(1)

		go func(qName string, handler SqsWorkerRegFunc) {
			defer wg.Done()

			w.Logger.Infof("Consuming queue '%s'", qName)
			for {
				if w.stopPolling {
					break
				}

				msgs, err := adapter.Consume(qName)
				if err != nil {
					w.Logger.Errorf("Consuming queue '%s' started with error: %v", qName, err)

					time.Sleep(RetryConsumeTimeoutMs * time.Millisecond)
					continue
				}

				for _, message := range msgs {
					w.Logger.Infof("Received a message from '%s'", qName)
					w.Logger.Debugf("Received message body: '%s'", aws.StringValue(message.Body))

					// Single thread processing. Adapters can be none thread safe!
					w.mutex.Lock()
					err = handler(message, w.Adapters)
					w.mutex.Unlock()

					if err != nil {
						w.Logger.Errorf("Queue '%s' failed to proceed the message '%s ", qName, message)

						return
					}

					err = adapter.DeleteMessage(qName, *message.ReceiptHandle)
					if err != nil {
						w.Logger.Errorf("Failed to delete the message '%s' from queue '%s'", message, qName)

						return
					}
				}

				w.Logger.Infof("Consuming queue '%s' stopped", qName)
			}
		}(queueName, handler)
	}

	w.waitChan = make(chan struct{})
	<-w.waitChan
	wg.Wait()
}

func (w *SqsEventsWorker) Stop() {
	w.Logger.Info("stop signal received! Graceful shutting down")

	w.stopPolling = true
	close(w.waitChan)
}
