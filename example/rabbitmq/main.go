package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/radianteam/framework"
	rmq_adapter "github.com/radianteam/framework/adapter/event/rabbitmq"
	"github.com/radianteam/framework/worker"
	rmq_worker "github.com/radianteam/framework/worker/event/rabbitmq"
	"github.com/radianteam/framework/worker/service/rest"
	"github.com/radianteam/framework/worker/task/job"
	"github.com/streadway/amqp"
)

var lastMessage string

func InitRabbitMqExchange(ctx context.Context, wc *worker.WorkerAdapters) error {
	rmqAdapter, _ := wc.Get("rmq")
	rmqAdapter.(*rmq_adapter.RabbitMqAdapter).DeclareQueue("test_queue", true)

	return nil
}

// RabbitMQ handler: reads queue and writes message body into lastMessage
func handlerRabbitMq(d *amqp.Delivery, wc *worker.WorkerAdapters) error {
	lastMessage = string(d.Body)

	return nil
}

// REST handler: reads lastMessage and returns text
func handlerRead(c *gin.Context, wc *worker.WorkerAdapters) {
	c.String(http.StatusOK, fmt.Sprintf("The last message: %s\n", lastMessage))
}

// REST handler: reads POST body and sends an event to RabbitMQ
func handlerSend(c *gin.Context, wc *worker.WorkerAdapters) {
	buff, _ := io.ReadAll(c.Request.Body)

	rmqAdapter, _ := wc.Get("rmq")
	rmqAdapter.(*rmq_adapter.RabbitMqAdapter).Publish("test_queue", buff)

	c.String(http.StatusOK, "")
}

func main() {
	// create a new microservice instance
	radian := framework.NewRadianMicroservice("main")

	// create init prejob to declare a queue
	initMqJob := job.NewTaskJob("init_mq", InitRabbitMqExchange)

	// create an adapter for rabbitmq
	adapterMqConfig := &rmq_adapter.RabbitMqConfig{Host: "rabbitmq", Port: 5672, Username: "example", Password: "pass", Exchange: ""}
	adapterMq := rmq_adapter.NewRabbitMqAdapter("rmq", adapterMqConfig)
	initMqJob.SetAdapter(adapterMq)

	// add prejob in the framework
	radian.AddPreJob(initMqJob)

	// create a new REST worker
	workerRestConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8088}
	workerRest := rest.NewRestServiceWorker("service_rest", workerRestConfig)

	// create routes to the worker
	workerRest.SetRoute("GET", "/", handlerRead)
	workerRest.SetRoute("POST", "/", handlerSend)

	// add the mq adapter to the worker
	workerRest.SetAdapter(adapterMq)

	// create a new RabbitMQ worker
	workerMqConfig := &rmq_worker.RabbitMqConfig{Host: "rabbitmq", Port: 5672, Username: "example", Password: "pass"}
	workerMq := rmq_worker.NewRabbitMqEventWorker("event_mq", workerMqConfig)

	// set handlers to the worker
	workerMq.SetEvent("test_queue", "test_queue", handlerRabbitMq)

	// append workers to the framework
	radian.AddWorker(workerRest)
	radian.AddWorker(workerMq)

	// run the workers
	radian.RunAll()
}
