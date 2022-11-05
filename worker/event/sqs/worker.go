package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	sqs_adapter "github.com/radianteam/framework/adapter/event/sqs"
	"github.com/radianteam/framework/worker"
	"sync"
)

type SqsWorkerRegFunc func(d *sqs.Message, wc *worker.WorkerAdapters) error

type SqsConfig struct {
	Host            string `json:"host,omitempty" config:"host,required"`
	Port            int16  `json:"port,omitempty" config:"port,required"`
	Region          string `json:"region,omitempty" config:"region,required"`
	AccessKeyId     string `json:"awsAccessKeyId,omitempty" config:"awsAccessKeyId,required"`
	SecretAccessKey string `json:"awsSecretAccessKey,omitempty" config:"awsSecretAccessKey,required"`
	SessionToken    string `json:"awsSessionToken,omitempty" config:"awsSessionToken,required"`
}

type SqsEventsWorker struct {
	*worker.BaseWorker

	config *SqsConfig

	mutex    sync.Mutex
	waitChan chan bool

	stopPolling bool

	handlers map[string]SqsWorkerRegFunc
}

func NewSqsEventsWorker(name string, config *SqsConfig) *SqsEventsWorker {
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

	adapterSqsConfig := &sqs_adapter.SqsConfig{
		Host:            "localstack",
		Port:            4566,
		AccessKeyId:     "test_key_id",
		SecretAccessKey: "test_secret_access_key",
		SessionToken:    "test_token",
		Region:          "us-east-2",
		PollingTime:     10,
	}
	adapter := sqs_adapter.NewSqsAdapter("sqs-consumer", adapterSqsConfig)
	err := adapter.Setup()
	if err != nil {
		w.Logger.Errorf("Failed to run '%s' worker during adapter configuration: %s", w.GetName(), err)

		return
	}

	w.waitChan = make(chan bool)
	for {
		if w.stopPolling {
			break
		}

		for queueName, handler := range w.handlers {
			wg.Add(1)

			go func(qName string, handler SqsWorkerRegFunc) {
				defer wg.Done()

				w.Logger.Infof("Consuming queue '%s'", qName)

				getQueueUrlInput := &sqs.GetQueueUrlInput{QueueName: aws.String(qName)}
				queueUrl, err := adapter.GetQueueUrl(getQueueUrlInput)
				if err != nil {
					w.Logger.Errorf("Failed to obtain url for queue '%s': %v", qName, err)

					return
				}

				receiveMessageInput := &sqs.ReceiveMessageInput{QueueUrl: aws.String(queueUrl)}
				msgs, err := adapter.Consume(receiveMessageInput)
				if err != nil {
					w.Logger.Errorf("Consuming queue '%s' started with error: %v", qName, err)

					return
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

					deleteMessageInput := &sqs.DeleteMessageInput{QueueUrl: aws.String(queueUrl), ReceiptHandle: message.ReceiptHandle}
					err = adapter.DeleteMessage(deleteMessageInput)
					if err != nil {
						w.Logger.Errorf("Failed to delete the message '%s' from queue '%s'", message, qName)

						return
					}
				}

				w.Logger.Infof("Consuming queue '%s' stopped", qName)
			}(queueName, handler)

			wg.Wait()
		}
	}
}

func (w *SqsEventsWorker) Stop() {
	w.Logger.Info("stop signal received! Graceful shutting down")

	w.stopPolling = true
	close(w.waitChan)
}
