# Example: simple REST service

1. [`Manual`](#1-manual)
2. [`Docker compose`](#2-docker-compose)

## 1 Manual

Create new folder. Create go.mod file inside with the following content:

``` go
module example

go 1.19

require (
	github.com/radianteam/framework v0.2.8
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
	// create a new framework instance
	radian := framework.NewRadianFramework()
}
```

After the instance create a configuration for a REST service and create a new REST service that listens all adresses and port 8088. Name this service as you wish ("service_rest" for example):

``` go
    // create a new REST worker
	workerConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8088}
	workerRest := rest.NewRestServiceWorker("service_rest", workerConfig)
```

OPTIONAL: create a database adapter for a sqlite3 database. Then append it to the REST worker:

``` go
    // create a database adapter
	dbConfig := &sqlx.SqlxConfig{Driver: "sqlite3", ConnectionString: "db.sqlite"}
	dbAdapter := sqlx.NewSqlxAdapter("db", dbConfig)

	//add the adapter to the worker
	workerRest.SetAdapter(dbAdapter)
```

Declare a handler for a GET request with a root path (function will be implemented later):

``` go
    // create a route to the worker
	workerRest.SetRoute("GET", "/", handlerMain)
```

Add the REST worker to the main framework instance:

``` go
    // append worker to the framework
	radian.AddWorker(workerRest)
```

Run the framework instance with the particular services:

``` go
    // run the worker
	radian.Run([]string{workerRest.GetName()})
```

Final step: declare and implement a handler function above the main fnction:

``` go
// REST handler function
func handlerMain(c *gin.Context, wc *worker.WorkerAdapters) {
	// extract the database adapter
	_, err := wc.Get("db")

	if err != nil {
		c.String(http.StatusBadRequest, "")
		return
	}

	// use the adapters and whatever you want

	// return standard gin results
	c.String(http.StatusOK, "Hello world!\n")
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

The answer has been received! If something goes wrong check [`main.go`](main.go) file or play with it in containers.

<br>

## 2 Docker compose

### 1 Clone the repository

```
git clone https://github.com/radianteam/framework.git
```
```
cd framework
```

### 2 Goto this folder

```
cd example/rest_simple
```


### 3 Run the application

```
docker-compose up -d
```

### 4 Make a request
Commands:
```
curl 127.0.0.1:8088/ 
```

Example
```
radian@radian:~$ curl 127.0.0.1:8088/                                   
Hello world!
```

### 5 Enjoy!

And don't forget to stop the application :)

```
docker-compose down
```
