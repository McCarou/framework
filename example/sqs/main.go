package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/radianteam/framework"
	sqs_adapter "github.com/radianteam/framework/adapter/event/sqs"
	"github.com/radianteam/framework/worker"
	sqs_worker "github.com/radianteam/framework/worker/event/sqs"
	"github.com/radianteam/framework/worker/service/rest"
	"io"
	"net/http"
)

const (
	inQueue    = "in"
	outQueue   = "out"
	sqsAdapter = "sqs-adapter"
)

func handlerRestIn(c *gin.Context, wc *worker.WorkerAdapters) {
	// receive message from POST request
	messageBytes, _ := io.ReadAll(c.Request.Body)
	messageString := string(messageBytes)

	// get sqs adapter from all running adapters
	adapter, _ := wc.Get(sqsAdapter)
	adapterSqs := adapter.(*sqs_adapter.SqsAdapter)

	// publish to the input queue
	adapterSqs.Publish(inQueue, messageString)
}

func fromInToOutQueueHandler(message *sqs.Message, wc *worker.WorkerAdapters) error {
	// get sqs adapter from all running adapters
	adapter, _ := wc.Get(sqsAdapter)
	adapterSqs := adapter.(*sqs_adapter.SqsAdapter)

	// publish to the output queue
	adapterSqs.Publish(outQueue, aws.StringValue(message.Body))

	return nil
}

func handlerRestOut(c *gin.Context, wc *worker.WorkerAdapters) {
	// get sqs adapter from all running adapters
	adapter, _ := wc.Get(sqsAdapter)
	adapterSqs := adapter.(*sqs_adapter.SqsAdapter)

	// read from the output queue
	result, _ := adapterSqs.Consume(outQueue)

	// remove message after consuming
	adapterSqs.DeleteMessage(outQueue, *result[0].ReceiptHandle)

	// return response
	c.String(http.StatusOK, aws.StringValue(result[0].Body))
}

func main() {
	// create a new framework instance
	radian := framework.NewRadianMicroservice("sqs-example")

	// setup sqs adapter
	adapterSqsConfig := &sqs_adapter.SqsConfig{
		Endpoint:            "http://localstack:4566",
		AccessKeyID:         "test_key_id",
		SecretAccessKey:     "test_secret_access_key",
		SessionToken:        "test_token",
		Region:              "us-east-2",
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     1,
		VisibilityTimeout:   1,
	}
	adapterSqs := sqs_adapter.NewSqsAdapter(sqsAdapter, adapterSqsConfig)
	adapterSqs.Setup()

	// create queue
	adapterSqs.CreateQueue(inQueue)
	adapterSqs.CreateQueue(outQueue)

	// setup rest worker
	restConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8080}
	workerRest := rest.NewRestServiceWorker("service_rest", restConfig)

	workerRest.SetRoute("POST", "/", handlerRestIn)
	workerRest.SetRoute("GET", "/", handlerRestOut)
	workerRest.SetAdapter(adapterSqs)

	// setup sqs worker
	workerSqs := sqs_worker.NewSqsEventsWorker("service_sqs", adapterSqsConfig)
	workerSqs.SetEvent(inQueue, fromInToOutQueueHandler)
	workerSqs.SetAdapter(adapterSqs)

	radian.AddWorker(workerSqs)
	radian.AddWorker(workerRest)

	// run the framework
	radian.RunAll()
}
