# Nanoservice framework Radian

The framework is designed to develop backend business applications based on SOA architecture.

DISCLAIMER: The project is at an early stage of development and is recommended for use with great caution. The project is open for general download due to the fact that the author needs it to support his current projects.

## Usage

#### 1 Create a new project and main.go file
#### 2 Put this code to the main.go

```
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/radianteam/framework"
	"github.com/radianteam/framework/context/repository/sqlx"
	"github.com/radianteam/framework/worker"
	"github.com/radianteam/framework/worker/service/rest"
)

// REST handler function
func handler_main(c *gin.Context, wc *worker.WorkerContexts) {
	// extract the database context
	_, err := wc.Get("db")

	if err != nil {
		c.String(http.StatusBadRequest, "")
		return
	}

	// use the contexts and whatever you want

	// return standard gin results
	c.String(http.StatusOK, "Hello world!\n")
}

func main() {
	// create a new framework instance
	radian := framework.NewRadianFramework()

	// create a new REST worker
	worker_config := &worker.WorkerConfig{ListenHost: "0.0.0.0", Port: 8088}
	worker_rest := rest.NewServiceRest(worker_config)

	// create a database context
	worker_context := sqlx.NewContextSqlx("sqlite3", "db.sqlite")

	//add the context to the worker
	worker_rest.AddContext("db", worker_context)

	// create a route to the worker
	worker_rest.SetRoute("GET", "/", handler_main)

	// append worker to the framework
	radian.AddWorker("service_rest", worker_rest)

	// run the worker
	radian.Run([]string{"service_rest"})
}
```

#### 3 Check

```
radian@radian:~$ curl 127.0.0.1:8088
Hello world!
```

#### 4 Enjoy!

You can download the test project [here](https://github.com/radianteam/framework-test "Radian Framework Test Project").