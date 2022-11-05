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
	"strconv"
)

const (
	inQueue    = "in"
	outQueue   = "out"
	sqsAdapter = "sqs-adapter"
)

func handlerRestIn(c *gin.Context, wc *worker.WorkerAdapters) {
	messageBytes, _ := io.ReadAll(c.Request.Body)
	messageString := string(messageBytes)

	adapter, _ := wc.Get(sqsAdapter)
	adapterSqs := adapter.(*sqs_adapter.SqsAdapter)

	getQueueUrlInput := &sqs.GetQueueUrlInput{QueueName: aws.String(inQueue)}
	queueUrl, _ := adapterSqs.GetQueueUrl(getQueueUrlInput)

	sendMessageInput := &sqs.SendMessageInput{QueueUrl: aws.String(queueUrl), MessageBody: aws.String(messageString)}
	_ = adapterSqs.Publish(sendMessageInput)
}

func fromInToOutQueueHandler(d *sqs.Message, wc *worker.WorkerAdapters) error {
	adapter, _ := wc.Get(sqsAdapter)

	adapterSqs := adapter.(*sqs_adapter.SqsAdapter)

	getQueueUrlInput := &sqs.GetQueueUrlInput{QueueName: aws.String(outQueue)}
	queueUrl, _ := adapterSqs.GetQueueUrl(getQueueUrlInput)

	sendMessageInput := &sqs.SendMessageInput{QueueUrl: aws.String(queueUrl), MessageBody: d.Body}
	_ = adapterSqs.Publish(sendMessageInput)

	return nil
}

func handlerRestOut(c *gin.Context, wc *worker.WorkerAdapters) {
	adapter, _ := wc.Get(sqsAdapter)

	adapterSqs := adapter.(*sqs_adapter.SqsAdapter)

	getQueueUrlInput := &sqs.GetQueueUrlInput{QueueName: aws.String(outQueue)}
	queueUrl, _ := adapterSqs.GetQueueUrl(getQueueUrlInput)

	recvMessageInput := &sqs.ReceiveMessageInput{QueueUrl: aws.String(queueUrl)}
	result, _ := adapterSqs.Consume(recvMessageInput)

	deleteMessageInput := &sqs.DeleteMessageInput{QueueUrl: aws.String(queueUrl), ReceiptHandle: result[0].ReceiptHandle}
	_ = adapterSqs.DeleteMessage(deleteMessageInput)

	c.String(http.StatusOK, aws.StringValue(result[0].Body))
}

func main() {
	adapterSqsConfig := &sqs_adapter.SqsConfig{
		Host:            "localstack",
		Port:            4566,
		AccessKeyId:     "test_key_id",
		SecretAccessKey: "test_secret_access_key",
		SessionToken:    "test_token",
		Region:          "us-east-2",
		PollingTime:     10,
	}
	adapterSqs := sqs_adapter.NewSqsAdapter(sqsAdapter, adapterSqsConfig)
	_ = adapterSqs.Setup()

	createInQueueInput := &sqs.CreateQueueInput{QueueName: aws.String(inQueue),
		Attributes: aws.StringMap(map[string]string{"ReceiveMessageWaitTimeSeconds": strconv.Itoa(adapterSqsConfig.PollingTime)})}
	_ = adapterSqs.CreateQueue(createInQueueInput)

	createOutQueueInput := &sqs.CreateQueueInput{QueueName: aws.String(outQueue),
		Attributes: aws.StringMap(map[string]string{"ReceiveMessageWaitTimeSeconds": strconv.Itoa(adapterSqsConfig.PollingTime)})}
	_ = adapterSqs.CreateQueue(createOutQueueInput)

	radian := framework.NewRadianFramework()

	restConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8080}
	work_est := rest.NewRestServiceWorker("service_rest", restConfig)

	work_est.SetRoute("POST", "/", handlerRestIn)
	work_est.SetRoute("GET", "/", handlerRestOut)
	work_est.SetAdapter(adapterSqs)

	workerConfig := &sqs_worker.SqsConfig{
		Host:            "localstack",
		Port:            4566,
		AccessKeyId:     "test_key_id",
		SecretAccessKey: "test_secret_access_key",
		SessionToken:    "test_token",
		Region:          "us-east-2",
	}
	workerSqs := sqs_worker.NewSqsEventsWorker("service_sqs", workerConfig)
	workerSqs.SetEvent(inQueue, fromInToOutQueueHandler)
	workerSqs.SetAdapter(adapterSqs)

	radian.AddWorker(workerSqs)
	radian.AddWorker(work_est)

	radian.Run([]string{"service_rest", "service_sqs"})
}
