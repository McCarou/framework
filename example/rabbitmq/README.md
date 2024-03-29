# Example: RabbitMQ

1. [`Manual`](#1-manual)
2. [`Docker compose`](#2-docker-compose)

## 1 Manual

Create new folder. Create go.mod file inside with the following content:

``` go
module example

go 1.19

require (
	github.com/radianteam/framework v0.3.0
)
```

This file declares a module and the framework requirement.

Create a file named main.go and define the package inside:

``` go
package main
```

Create a main function and create an instance of the radian framework:

``` go
func main() {
	// create a new microservice instance
	radian := framework.NewRadianMicroservice()
}
```

Above the main function declare some handlers. First of all declare a string for the last message

``` go
var lastMessage string
```

Then implement a job to declare a test queue

``` go
func InitRabbitMqExchange(ctx context.Context, wc *worker.WorkerAdapters) error {
	rmqAdapter, _ := wc.Get("rmq")
	rmqAdapter.(*rmq_adapter.RabbitMqAdapter).DeclareQueue("test_queue", true)

	return nil
}
```

Next implement a RabbitMQ handler that receives messages and store them to the variable

``` go
// RabbitMQ handler: reads queue and writes message body into lastMessage
type HandlerRabbitMq struct {
	rmq_worker.RabbitMqEventHandler
}

func (h *HandlerRabbitMq) Handle() error {
	lastMessage = string(h.MqMessage.Body)

	return nil
}
```

Finally implement 2 rest handlers: one for reading the last message variable

``` go
// REST handler: reads lastMessage and returns text
type HandlerRead struct {
	rest.RestServiceHandler
}

func (h *HandlerRead) Handle() error {
	h.GinContext.String(http.StatusOK, fmt.Sprintf("The last message: %s\n", lastMessage))

	return nil
}
```

And one for sending messages into RabbitMQ

``` go
// REST handler: reads POST body and send an event to RabbitMQ
type HandlerSend struct {
	rest.RestServiceHandler
}

func (h *HandlerSend) Handle() error {
	buff, _ := io.ReadAll(h.GinContext.Request.Body)

	rmqAdapter, _ := h.Adapters.Get("rmq")
	rmqAdapter.(*rmq_adapter.RabbitMqAdapter).Publish("test_queue", buff)

	h.GinContext.String(http.StatusOK, "")

	return nil
}
```

In the main function after the instance create a new prejob. Name this service as you with ("init_mq" for example). Then declare rabbitmq server credentials, set the adapter in the prejob and add prejob in the framework:

``` go
    // create init prejob to declare a queue
	initMqJob := job.NewTaskJob("init_mq", InitRabbitMqExchange)

	// create an adapter for rabbitmq
	adapterMqConfig := &rmq_adapter.RabbitMqConfig{Host: "rabbitmq", Port: 5672, Username: "example", Password: "pass", Exchange: ""}
	adapterMq := rmq_adapter.NewRabbitMqAdapter("rmq", adapterMqConfig)
	initMqJob.SetAdapter(adapterMq)

	// add prejob in the framework
	radian.AddPreJob(initMqJob)
```

After create a new REST service that listens all adresses and port 8088 and set handlers to handle requests. Name this service as you with ("service_rest" for example):

``` go
    // create a new REST worker
	workerRestConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8088}
	workerRest := rest.NewRestServiceWorker("service_rest", workerRestConfig)

	// create routes to the worker
	workerRest.SetRoute("GET", "/", &HandlerRead{})
	workerRest.SetRoute("POST", "/", &HandlerSend{})
```

Append the RabbitMQ adapter from the prejob. You can reuse it because the prejob doesn't ned it anymore:

``` go
	// add the mq adapter to the worker
	workerRest.SetAdapter(adapterMq)
```

Create a new RabbitMQ worker and add the handler to handle messages from the test queue

``` go
	// create a new RabbitMQ worker
	workerMqConfig := &rmq_worker.RabbitMqConfig{Host: "rabbitmq", Port: 5672, Username: "example", Password: "pass"}
	workerMq := rmq_worker.NewRabbitMqEventWorker("event_mq", workerMqConfig)

	// set handlers to the worker
	workerMq.SetEvent("test_queue", "test_queue", &HandlerRabbitMq{})
```

Add the workers to the main framework instance:

``` go
    // append workers to the framework
	radian.AddWorker(workerRest)
	radian.AddWorker(workerMq)
```

Run the framework instance with workers and tasks:

``` go
    // run the workers
	radian.RunAll()
```
<br>

Now run the following command to download all requriments and prepare the application to start:

```
go mod tidy -compat=1.19
```

Wait for the requirements download and then run the app:

```
go run main.go
```

Then run a test command to send a new message. Like:
```
curl -X POST -d "Hello world" 127.0.0.1:8088/
```

The message will be send and consumed by the RabbitMQ handler:

``` json
{"level":"info","msg":"Received a message from test_queue with key test_queue","time":"2022-11-01T19:49:06Z","worker":"event_mq"}
```

Then check the last message:

```
curl 127.0.0.1:8088/
```

Example
```
radian@radian:~$ curl 127.0.0.1:8088/                        
The last message: Hello world
```

The message has been consumed! If something goes wrong you can check [`main.go`](main.go) file or play with it in containers.
<br><br>

## 2 Docker compose

WARNING: you must have docker and docker-compose installed on your system. Use [`this instruction`](https://docs.docker.com/compose/install/) if you don't have it.

### 1 Clone the repository

```
git clone https://github.com/radianteam/framework.git
```
```
cd framework
```

### 2 Goto this folder

```
cd example/rabbitmq
```


### 3 Run the application

```
docker-compose up -d
```

### 4 Make an event
Commands:
```
curl -X POST -d "Hello world" 127.0.0.1:8088/
```

### 5 Read the last RabbitMQ message
Commands:
```
curl 127.0.0.1:8088/
```

Example
```
radian@radian:~$ curl 127.0.0.1:8088/                        
The last message: Hello world
```

### 6 Enjoy!

And don't forget to stop the application :)

```
docker-compose down
```
