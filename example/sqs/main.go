package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/radianteam/framework"
	sqs_adapter "github.com/radianteam/framework/adapter/event/sqs"
	sqs_worker "github.com/radianteam/framework/worker/event/sqs"
	"github.com/radianteam/framework/worker/service/rest"
)

const (
	inQueue    = "in"
	outQueue   = "out"
	sqsAdapter = "sqs-adapter"
)

type HandlerRestIn struct {
	rest.RestServiceHandler
}

func (h *HandlerRestIn) Handle() error {
	// receive message from POST request
	messageBytes, _ := io.ReadAll(h.GinContext.Request.Body)
	messageString := string(messageBytes)

	// get sqs adapter from all running adapters
	adapter, _ := h.Adapters.Get(sqsAdapter)
	adapterSqs := adapter.(*sqs_adapter.AwsSqsAdapter)

	// publish to the input queue
	adapterSqs.PublishQueue(inQueue, messageString)

	return nil
}

type QueueHandler struct {
	sqs_worker.AwsSqsEventHandler
}

func (h *QueueHandler) Handle() error {
	// get sqs adapter from all running adapters
	adapter, _ := h.Adapters.Get(sqsAdapter)
	adapterSqs := adapter.(*sqs_adapter.AwsSqsAdapter)

	// publish to the output queue
	adapterSqs.PublishQueue(outQueue, aws.StringValue(h.SqsMessage.Body))

	return nil
}

type HandlerRestOut struct {
	rest.RestServiceHandler
}

func (h *HandlerRestOut) Handle() error {
	// get sqs adapter from all running adapters
	adapter, _ := h.Adapters.Get(sqsAdapter)
	adapterSqs := adapter.(*sqs_adapter.AwsSqsAdapter)

	// read from the output queue
	result, _ := adapterSqs.Consume(outQueue)

	// remove message after consuming
	adapterSqs.DeleteMessage(outQueue, *result[0].ReceiptHandle)

	// return response
	h.GinContext.String(http.StatusOK, fmt.Sprintf("%s\n", aws.StringValue(result[0].Body)))

	return nil
}

func main() {
	// create a new framework instance
	radian := framework.NewRadianMicroservice("sqs-example")

	// setup sqs adapter
	adapterSqsConfig := &sqs_adapter.AwsSqsConfig{
		Endpoint:            "http://localstack:4566",
		AccessKeyID:         "test_key_id",
		SecretAccessKey:     "test_secret_access_key",
		SessionToken:        "test_token",
		Region:              "us-east-2",
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     1,
		VisibilityTimeout:   1,
	}
	adapterSqs := sqs_adapter.NewAwsSqsAdapter(sqsAdapter, adapterSqsConfig)
	adapterSqs.Setup()

	// create queue
	adapterSqs.CreateQueue(inQueue)
	adapterSqs.CreateQueue(outQueue)

	// setup rest worker
	restConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8080}
	workerRest := rest.NewRestServiceWorker("service_rest", restConfig)

	// setup routes for workers
	workerRest.SetRoute("POST", "/", &HandlerRestIn{})
	workerRest.SetRoute("GET", "/", &HandlerRestOut{})

	// set adapter to the worker
	workerRest.SetAdapter(adapterSqs)

	// setup sqs worker
	workerSqs := sqs_worker.NewAwsSqsEventsWorker("service_sqs", adapterSqsConfig)
	workerSqs.SetEvent(inQueue, &QueueHandler{})
	workerSqs.SetAdapter(adapterSqs)

	// add workers
	radian.AddWorker(workerSqs)
	radian.AddWorker(workerRest)

	// run the framework
	radian.RunAll()
}
