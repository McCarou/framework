# Example monitoring for the Radian framework

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

After the instance create a configuration for a REST service and create a new REST service that listens all adresses and port 8088. Name this service as you with ("service_rest" for example):

``` go
    // create a new REST worker
	workerConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8088}
	workerRest := rest.NewRestServiceWorker("service_rest", workerConfig)
```

Create a monitoring worker on another port:

``` go
    // create a new mnitoring worker
	prometheusConfig := &monitoring.MonitoringServiceConfig{Listen: "0.0.0.0", Port: 8087}
	workerPrometheus := monitoring.NewMonitoringServiceWorker("service_monitoring", prometheusConfig)
```

Declare a handler for a GET request with a root path (function will be implemented later):

``` go
    // create a route to the worker
	workerRest.SetRoute("GET", "/", &MainHandler{})
```

Enable monitoring for the worker:

``` go
    // enable metrics collecting
    workerRest.SetMonitoring(true)
```

Add the REST worker and the monitoring worker to the main framework instance:

``` go
    // append workers to the framework
	radian.AddWorker(workerRest)
	radian.AddWorker(workerPrometheus)
```

Run the framework instance with the particular services:

``` go
    // run the workers
	radian.RunAll()
```

Final step: declare and implement a handler function above the main fnction:

``` go
// REST handler struct
type MainHandler struct {
	rest.RestServiceHandler
}

// REST handler function
func (h *MainHandler) Handle() error {
	// return standard gin results
	h.GinContext.String(http.StatusOK, "Hello world!\n")

	return nil
}
```

This function extracts the database adapter and return a response to a client. Pretty useless but OK for the example.
<br>

Now run the following command to download all requriments and prepare the application to start:

```
go mod tidy
```

Wait for the requirements download and then run the app:

```
go run main.go
```

Then run a test command. Like:
```
curl 127.0.0.1:8088/ 
```

Example
```
radian@radian:~$ curl 127.0.0.1:8088/                                   
Hello world!
```

Check metrics
Commands:
```
curl 127.0.0.1:8087/metrics
```

Example
```
radian@radian:~$ curl 127.0.0.1:8087/metrics
...
promhttp_metric_handler_requests_total{code="200"} 0
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP rest_worker_total_requests Total requests of the rest worker
# TYPE rest_worker_total_requests counter
rest_worker_total_requests{code="200",method="GET",url="/",worker_name="service_rest"} 1
```

The answer has been received! If something goes wrong you can check [`main.go`](main.go) file or play with it in containers.
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
cd example/monitoring
```


### 3 Run the application

```
docker-compose up -d
```

### 4 Make some requests
Commands:
```
curl 127.0.0.1:8088/ 
```
```
curl 127.0.0.1:8088/absent
```

Example
```
radian@radian:~$ curl 127.0.0.1:8088/                                   
Hello world!
radian@radian:~$ curl 127.0.0.1:8088/absent
404 page not found
```

### 5 Check metrics
Commands:
```
curl 127.0.0.1:8087/metrics
```

Example
```
radian@radian:~$ curl 127.0.0.1:8087/metrics
...
promhttp_metric_handler_requests_total{code="200"} 0
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
# HELP rest_worker_total_requests Total requests of the rest worker
# TYPE rest_worker_total_requests counter
rest_worker_total_requests{code="200",method="GET",url="/",worker_name="service_rest"} 1
rest_worker_total_requests{code="404",method="GET",url="/absent",worker_name="service_rest"} 1
```

### 6 Enjoy!

Compose also runs poller to send periodically requests to the framework application and prometheus + grafana services.
To check metrics in prometheus open in browser:
```
http://localhost:9000/graph?g0.expr=rest_worker_total_requests&g0.tab=1&g0.stacked=0&g0.show_exemplars=0&g0.range_input=1h
```

Also these metrics are exported into grafana (login: admin, password: foobar):
```
http://127.0.0.1:3000/d/yVAO0HH4k/radian-rest-worker-dashboard?orgId=1&refresh=10s
```

And don't forget to stop the application :)

```
docker-compose down
```
