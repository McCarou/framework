package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/radianteam/framework"
	rmq_adapter "github.com/radianteam/framework/adapter/event/rabbitmq"
	rmq_worker "github.com/radianteam/framework/worker/event/rabbitmq"
	"github.com/radianteam/framework/worker/service/rest"
	"github.com/radianteam/framework/worker/task/job"
)

var lastMessage string

type HandlerInitMq struct {
	job.TaskJobHandler
}

func (h *HandlerInitMq) Handle() error {
	rmqAdapter, _ := h.Adapters.Get("rmq")
	rmqAdapter.(*rmq_adapter.RabbitMqAdapter).DeclareQueue("test_queue", true)

	return nil
}

type HandlerRabbitMq struct {
	rmq_worker.RabbitMqEventHandler
}

// RabbitMQ handler: reads queue and writes message body into lastMessage
func (h *HandlerRabbitMq) Handle() error {
	lastMessage = string(h.MqMessage.Body)

	return nil
}

type HandlerRead struct {
	rest.RestServiceHandler
}

// REST handler: reads lastMessage and returns text
func (h *HandlerRead) Handle() error {
	h.GinContext.String(http.StatusOK, fmt.Sprintf("The last message: %s\n", lastMessage))

	return nil
}

type HandlerSend struct {
	rest.RestServiceHandler
}

// REST handler: reads POST body and sends an event to RabbitMQ
func (h *HandlerSend) Handle() error {
	buff, _ := io.ReadAll(h.GinContext.Request.Body)

	rmqAdapter, _ := h.Adapters.Get("rmq")
	rmqAdapter.(*rmq_adapter.RabbitMqAdapter).Publish("test_queue", buff)

	h.GinContext.String(http.StatusOK, "")

	return nil
}

func main() {
	// create a new microservice instance
	radian := framework.NewRadianMicroservice("main")

	// create init prejob to declare a queue
	initMqJob := job.NewTaskJob("init_mq", &HandlerInitMq{})

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
	workerRest.SetRoute("GET", "/", &HandlerRead{})
	workerRest.SetRoute("POST", "/", &HandlerSend{})

	// add the mq adapter to the worker
	workerRest.SetAdapter(adapterMq)

	// create a new RabbitMQ worker
	workerMqConfig := &rmq_worker.RabbitMqConfig{Host: "rabbitmq", Port: 5672, Username: "example", Password: "pass"}
	workerMq := rmq_worker.NewRabbitMqEventWorker("event_mq", workerMqConfig)

	// set handlers to the worker
	workerMq.SetEvent("test_queue", "test_queue", &HandlerRabbitMq{})

	// append workers to the framework
	radian.AddWorker(workerRest)
	radian.AddWorker(workerMq)

	// run the workers
	radian.RunAll()
}
