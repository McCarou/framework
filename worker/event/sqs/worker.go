package sqs

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	sqs_adapter "github.com/radianteam/framework/adapter/event/sqs"
	"github.com/radianteam/framework/worker"
)

const RetryConsumeTimeoutMs = 10000

type AwsSqsEventsWorker struct {
	*worker.BaseWorker

	config *sqs_adapter.AwsSqsConfig

	mutex       sync.Mutex
	waitChan    chan struct{}
	stopPolling bool

	handlers map[string]AwsSqsEventHandlerInterface
}

func NewAwsSqsEventsWorker(name string, config *sqs_adapter.AwsSqsConfig) *AwsSqsEventsWorker {
	handlers := make(map[string]AwsSqsEventHandlerInterface)

	return &AwsSqsEventsWorker{BaseWorker: worker.NewBaseWorker(name), config: config, handlers: handlers}
}

func (w *AwsSqsEventsWorker) SetEvent(queue string, handler AwsSqsEventHandlerInterface) {
	w.handlers[queue] = handler
}

func (w *AwsSqsEventsWorker) Setup() {
	w.Logger.Info("Setting up Sqs Events")
}

func (w *AwsSqsEventsWorker) Run() {
	w.Logger.Info("Running Sqs Events Worker")

	wg := sync.WaitGroup{}

	adapter := sqs_adapter.NewAwsSqsAdapter("sqs-consumer", w.config)
	err := adapter.Setup()
	if err != nil {
		w.Logger.Errorf("Failed to run '%s' worker during adapter configuration: %s", w.GetName(), err)

		return
	}

	for queueName, handler := range w.handlers {
		wg.Add(1)

		go func(qName string, handler AwsSqsEventHandlerInterface) {
			defer wg.Done()

			handler.SetLogger(w.Logger.WithField("queue", qName))
			handler.SetAdapters(w.Adapters)

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
					handler.SetSqsMessage(message)
					err = handler.Handle()
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

				w.Logger.Debugf("Consuming queue '%s' stopped", qName)
			}
		}(queueName, handler)
	}

	w.waitChan = make(chan struct{})
	<-w.waitChan
	wg.Wait()
}

func (w *AwsSqsEventsWorker) Stop() {
	w.Logger.Info("stop signal received! Graceful shutting down")

	w.stopPolling = true
	close(w.waitChan)
}
