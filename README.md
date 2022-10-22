# Nanoservice framework Radian

The framework is designed to develop backend business applications based on SOA architecture. The framework supported microservices and monolith architecture.

DISCLAIMER: The project is at an early stage of development and is recommended for use with great caution. The project is open for general download due to the fact that the author needs it to support his current projects.

# Menu

1. [`Framework`](#1-framework)
	1. [`Quick start`](#1-quick-start)
	2. [`Basic information`](#2-basic-information)
	3. [`Four types of interaction`](#3-four-types-of-interaction)
	4. [`Workers and adapters`](#4-workers-and-adapters)
2. [`Supported workers`](#2-supported-workers)
	1. [`Services`](#1-services)
		1. [`REST`](#1-rest)
		2. [`GRPC`](#2-grpc)
	2. [`Events`](#2-events)
		1. [`RabbitMQ`](#1-rabbitmq)
	3. [`Tasks`](#3-tasks)
		1. [`Schedule`](#1-schedule)
		2. [`Jobs`](#2-jobs)
	4. [`Utility`](#4-utility)
		1. [`Monitoring`](#1-monitoring)
3. [`Supported adapters`](#3-supported-adapters)
	1. [`Utility`](#1-utility)
		1. [`Configuration`](#1-configuration)
	2. [`Storage`](#2-storage)
		1. [`Sqlx`](#1-sqlx)
		2. [`MongoDB`](#2-mongodb)
		3. [`ArangoDB`](#3-arangodb)
	3. [`Events`](#3-events)
		1. [`RabbitMQ`](#1-rabbitmq)
	4. [`Auth`](#4-auth)
		1. [`OIDC`](#1-oidc)
4. [`Project organization`](#4-project-organization)
	1. [`Main code`](#1-main-code)
	2. [`Workers`](#2-workers)
	3. [`Work with data`](#3-work-with-data)
	4. [`Custom adapters`](#4-custom-adapters)
	5. [`Extra`](#5-extra)
	6. [`Framework lifecycle`](#6-framework-lifecycle)
5. [`External links`](#5-external-links)
	1. [`Examples`](#1-examples)
		1. [`Simple REST service`](example/simple_rest)
		2. [`REST service with monitoring`](example/monitoring)

# Documentation
## 1 Framework
### 1 Quick start

Create a new project and main.go file and put this code to the main.go

``` go
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

This file contains a REST worker with a handler function which gets database adapter and return status code with payload "Hello world!"
<br><br>
Main function:
1. creates a framework instance
2. creates a database adapter connected to a file database
3. creates a new REST worker
4. sets the handler to the REST worker
5. runs the worker

Run the following command

```
curl 127.0.0.1:8088
```

Application will show the response

```
Hello world!
```

Also check [`examples`](example).

<br>

### 2 Basic information
Radian framework is developed to support most of the cloud native patterns and bring new style to create tech projects from business requests and archirecture based on components and interactions. In the framework they are named like workers and adapters. Also the framework is based on four types interaction concept. Author had a lot of experience and developed this framework to help small teams to organize their code and architecture and for hude projects and distributed teams prove a solution to work together without pain (like how to share protocols over teams, how to build microservices and run monolith, how to deal with informational security departments, etc).
<br><br>
Advantages of using this framwork:
- run application both as monolith or as microservices
- boost code organization for distributed teams
- use 4 interaction paradigm and manage your teams better
- build services with simple reusable blocks
- easy advance framework with new adapters
- have predictable outcomes and stay clear to management teams

Disadvantages:
- the framework is at an early stage of development (but several companies already use it and are happy!)
- the project is supported by only one developer
- you have your own better solution or another requirements
<br><br>

### 3 Four types of interaction
Not implemented
### 4 Workers and adapters
Not implemented
## 2 Supported workers
### 1 Services
#### 1 REST
Not implemented
#### 2 GRPC
Not implemented
### 2 Events
#### 1 RabbitMQ
Not implemented
### 3 Tasks
#### 1 Schedule
Not implemented
#### 2 Jobs
Not implemented
### 4 Utility
#### 1 Monitoring
Not implemented
## 3 Supported adapters
### 1 Utility
#### 1 Configuration
Not implemented
### 2 Storage
#### 1 Sqlx
Not implemented
#### 2 MongoDB
Not implemented
#### 3 ArangoDB
Not implemented
### 3 Events
#### 1 RabbitMQ
Not implemented
### 4 Auth
#### 1 OIDC
Not implemented
## 4 Project organization
### 1 Main code
Not implemented
### 2 Workers
Not implemented
### 3 Work with data
Not implemented
### 4 Custom adapters
Not implemented
### 5 Extra
Not implemented
### 6 Framework lifecycle
Not implemented
## 5 External links
### 1 Examples
#### 1 Simple REST service
Not implemented
#### 2 REST service with monitoring (prometheus metrics)
Not implemented
