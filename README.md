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
	sqlx "github.com/radianteam/framework/adapter/storage/sqlx"
	"github.com/radianteam/framework/worker"
	rest "github.com/radianteam/framework/worker/service/rest"
)

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

func main() {
	// create a new framework instance
	radian := framework.NewRadianFramework()

	// create a new REST worker
	workerConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8088}
	workerRest := rest.NewRestServiceWorker("service_rest", workerConfig)

	// create a database adapter
	dbConfig := &sqlx.SqlxConfig{Driver: "sqlite3", ConnectionString: "db.sqlite"}
	dbAdapter := sqlx.NewSqlxAdapter("db", dbConfig)

	//add the adapter to the worker
	workerRest.SetAdapter(dbAdapter)

	// create a route to the worker
	workerRest.SetRoute("GET", "/", handlerMain)

	// append worker to the framework
	radian.AddWorker(workerRest)

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

You can download the example project [here](https://github.com/radianteam/framework-example "Radian Framework Example Project").