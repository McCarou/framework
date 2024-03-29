# Example: simple SQS service

1. [`Manual`](#1-manual)
2. [`Docker compose`](#2-docker-compose)

### 1 Manual

Create a new folder. Create go.mod file inside with the following content:

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

Import all required dependencies

``` go
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
```

Declare required constants and a proxy variable

``` go
var lastMessage string

const (
	queueName  = "testqueue"
	sqsAdapter = "sqs-adapter"
)
```

Create a main function and create an instance of the radian framework:

``` go
func main() {
	// create a new micrservice instance
	radian := framework.NewRadianMicroservice()
}
```

After the instance create and setup new SQS adapter:

``` go
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
		Queue:               "testqueue",
	}
	adapterSqs := sqs_adapter.NewAwsSqsAdapter(sqsAdapter, adapterSqsConfig)
	adapterSqs.Setup()
```

Create a queue

``` go
    // create queue
	adapterSqs.CreateQueue(queueName)
```

Create a configuration for a REST service and create a new REST service that listens all adresses and port 8088. Name this service as you wish ("service_rest" for example):

``` go
    // setup rest worker
	restConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8080}
	workerRest := rest.NewRestServiceWorker("service_rest", restConfig)
```

Declare a handler for a GET request with a root path (function will be implemented later):

``` go
    // setup routes for workers
	workerRest.SetRoute("POST", "/", &HandlerRestIn{})
	workerRest.SetRoute("GET", "/", &HandlerRestOut{})
```

Add the adapter to the worker:
``` go
    // set adapter to the worker
	workerRest.SetAdapter(adapterSqs)
```

Setup AWS SQS worker:
``` go
    // setup sqs worker
	workerSqs := sqs_worker.NewAwsSqsEventsWorker("service_sqs", adapterSqsConfig)
	workerSqs.SetEvent(queueName, &QueueHandler{})
	workerSqs.SetAdapter(adapterSqs)
```

Add both SQS and REST workers to the main framework instance:

``` go
    // add workers
	radian.AddWorker(workerSqs)
	radian.AddWorker(workerRest)
```

Run the framework instance with the particular services:

``` go
    // run the worker
	radian.RunAll()
```

Declare and implement handlers functions above the main function.
The first REST handler receives a message sent by POST request and publishes it to the input queue.

``` go
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
	adapterSqs.PublishQueue(queueName, messageString)

	return nil
}
```

The second SQS handler extracts the message from the input queue and saves it to the proxy variable.

``` go
type QueueHandler struct {
	sqs_worker.AwsSqsEventHandler
}

func (h *QueueHandler) Handle() error {
	// save the message
	lastMessage = aws.StringValue(h.SqsMessage.Body)

	return nil
}
```

The third REST handler receives GET request and extracts the message from the proxy variable, then returns it in the response.
``` go
type HandlerRestOut struct {
	rest.RestServiceHandler
}

func (h *HandlerRestOut) Handle() error {
	h.GinContext.String(http.StatusOK, fmt.Sprintf("The last message: %s\n", lastMessage))

	return nil
}
```

<br>

Create `Dockerfile` and paste the following content:
```dockerfile
FROM golang:1.19 AS builder
WORKDIR /app
COPY ./ ./
RUN go mod tidy -compat=1.19
RUN go build -o app ./example/sqs

FROM ubuntu:latest AS app
WORKDIR /app
COPY --from=builder /app/app ./
CMD ["./app"]

```

#### Setting up SQS

Now we need to launch SQS itself with our application. We will use [localstack](https://github.com/localstack/localstack) to have our own instance of Amazon SQS.

Create `docker-compose.yml` file and paste the following content:
```yaml
version: "3.8"

services:
  app:
    build:
      context: ./../../
      dockerfile: example/sqs/Dockerfile
      target: app
    restart: "no"
    ports:
      - "8080:8080"
      - "4567:4566"
    depends_on:
      localstack:
        condition: service_healthy
  localstack:
    container_name: "${LOCALSTACK_DOCKER_NAME-localstack_main}"
    image: localstack/localstack
    ports:
      - "127.0.0.1:4566:4566"            # LocalStack Gateway
      - "127.0.0.1:4510-4559:4510-4559"  # external services port range
    environment:
      - DOCKER_HOST=unix:///var/run/docker.sock
    healthcheck:
      test: [ "CMD", "curl", "-f", "http://localhost:4566" ]
      interval: 10s
      timeout: 5s
      retries: 5

```

Now run the following command to download all requirements and prepare the application to start:

```bash
docker-compose up -d
```

Make a request to put a message to the input queue
```shell
curl -X POST http://localhost:8080/ -H "Content-Type: application/text" -d "HellO!"
```

Make a request to get the message from the output queue
Commands:
```shell
curl -X GET http://localhost:8080/
```

Example:
```
curl -X GET http://localhost:8080/
HellO!
```

Enjoy!

And also don't forget to stop the application :)

```shell
docker-compose down
```

<br>

## 2 Docker compose

WARNING: you must have docker and docker-compose installed on your system. Use [`this instruction`](https://docs.docker.com/compose/install/) if you don't have it.

### 1 Clone the repository

```shell
git clone https://github.com/radianteam/framework.git
cd framework
```

### 2 Goto this folder

```shell
cd example/sqs
```


### 3 Run the application

```shell
docker-compose up -d
```

### 4 Make a request to put a message to the input queue
Commands:
```shell
curl -X POST -H "Content-Type: application/text" -d "HellO" http://localhost:8080/
```

### 5 Make a request to get the message from the output queue
Commands:
```shell
curl -X GET http://localhost:8080/
```

Example:
```
curl -X GET http://localhost:8080/
HellO
```

### 6 Enjoy!

And also don't forget to stop the application :)

```shell
docker-compose down
```
