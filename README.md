# Nanoservice framework Radian

The framework is designed to develop backend business applications based on SOA architecture. The framework supported microservices and monolith architecture.

DISCLAIMER: The project is at an early stage of development and is recommended for use with great caution. The project is open for general download due to the fact that the author needs it to support his current projects.

# Menu

1. [`Framework`](#1-framework)
	1. [`Quick start`](#1-quick-start)
	2. [`Basic information`](#2-basic-information)
	3. [`Four types of interaction`](#3-four-types-of-interaction)
	4. [`Jobs`](#4-jobs)
	4. [`Workers and adapters`](#5-workers-and-adapters)
2. [`Supported workers`](#2-supported-workers)
3. [`Supported adapters`](#3-supported-adapters)
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
Radian framework is developed to support most of the cloud native patterns and bring new style to create tech projects from business requests and architecture based on components and interactions. In the framework they are named like workers and adapters. Also the framework is based on four types interaction concept. Author had a lot of experience and developed this framework to help small teams to organize their code and architecture and for huge projects and distributed teams prove a solution to work together without pain (like how to share protocols over teams, how to build microservices and run monolith, how to deal with informational security departments, etc).
<br><br>
Advantages of using this framework:
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

Radian uses "Four types of interaction" concept. The concept explains that any business backend application can be designed with this interactions:
1. Service interaction or request response. In this type a server listens for incoming connections and make clients wait during the response preparation. Features:
	- passive interaction: service waits for requests
	- sync interaction: clients wait while server prepares a response
	- answer guarantee or error: the are no promises or queues to sent a response as a callback. Client receives an answer immediately
	- stateless: every request has no any state. Clients must provide additional information about sessions
	- strong dependency: clients must know server protocols and their address and must control any request fault
	- consequent process: clients must wait for the previous request completion before starting the next
	- request retries: if fails request side must retry a request and control fault tolerance policies
	- can have DoS situation: if server is overflew with request it might not handle any more requests
Examples: REST or GRPC requests, HTTP, file download, etc.
2. Event interaction or producer-consumer. Events are async and provide decouple pattern. Features:
	- passive interaction: service waits for requests
	- async interaction: clients don't wait, just put a message in a queue and forget about it
	- answer not guarantee: the is a case when nobody listens to you messages
	- unknown amount of listeners: service doesn't who and when will read the messages
	- stateless: every queue message has no any state.
	- ready for mass requests: client can push a lot of messages and never mind about their handling
	- avoid DoS situation: all request are put in a queue. There can be queue overflow but in general the consumer will handle messages one by one with the maximum speed.
Examples: message broker queues (rabbitmq, kafka, sqs), message broadcast, log collecting, async tasks, etc.
3. Periodic interaction or tasks. A handler can setup periodic logic and run it in the particular time moment or after timeout. Features:
	- active interaction: service actively call functions to proceed
	- internal interaction: service doesn't interact in general. This is an internal process
	- fault tolerance ready: the common case to use this type of interaction is to restore software from unusual conditions or check some signals to react
	- schedule ready: service plan to run some tasks
	- wait or not: during processing task when service want to run another it can wait for the previous task completion or run another task with no waiting.
Examples: delayed tasks, fault restoration, periodic checks, status update, etc.
4. Permanent interaction or threads. This is non-stop processes inside microservices to make some internal logic or interact with local component or signal sources. Features:
	- run permanently: the is neither active nor passive interaction. This isn't interaction in general
	- restart policies: if errors a permanent thread can be restarted or left down
	- constant connection: perfect for listening some hardware signals or keeping connection to old style system with socket interaction
	- interaction through another interaction: permanent tasks can init any other interaction (make requests or put a message in a queue) to interract with other services
Examples: connect to financial market protocols, listen to hardware signal sources, etc.

Microservices combining these interactions can provide solutions with any complexity. They can have several instances of one interaction in one microservice. For example, a REST service and a GRPC service in the same time. And their name is NANOSERVICES. Framework can run these nanoservices together like a monolith or devops teams can tune them to run separately and have fine grained control for more security or predictable loading.
<br><br>

### 4 Jobs

Framework implements jobs that can be run before or after the main application loop. They are called pretasks and posttasks.
Pretasks can be used for:
- getting auth keys
- downloading session certs
- initialize something like queues and other stuff you may need before starting.

Posttasks are for:
- revoking auth session
- notice other processes about stopping
- deleting session data and queues

This tasks are not for migrations! You can use them for it but it is better to do migrations by devops. The only way to use tasks for migration or data seed is when you run your application as monolith.

<br><br>

### 5 Workers and adapters
Not implemented
<br><br>

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
