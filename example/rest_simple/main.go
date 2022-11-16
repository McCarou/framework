package main

import (
	"net/http"

	"github.com/radianteam/framework"
	sqlx "github.com/radianteam/framework/adapter/storage/sqlx"
	rest "github.com/radianteam/framework/worker/service/rest"
)

type MainHandler struct {
	rest.RestServiceHandler
}

// REST handler function
func (h *MainHandler) Handle() error {
	// extract the database adapter
	_, err := h.Adapters.Get("db")

	if err != nil {
		h.GinContext.String(http.StatusBadRequest, "")
		return err
	}

	// use the adapters and whatever you want

	// return standard gin results
	h.GinContext.String(http.StatusOK, "Hello world!\n")

	return nil
}

func main() {
	// create a new microservice instance
	radian := framework.NewRadianMicroservice("main")

	// create a new REST worker
	workerConfig := &rest.RestConfig{Listen: "0.0.0.0", Port: 8088}
	workerRest := rest.NewRestServiceWorker("service_rest", workerConfig)

	// create a database adapter
	dbConfig := &sqlx.SqlxConfig{Driver: "sqlite3", ConnectionString: "db.sqlite"}
	dbAdapter := sqlx.NewSqlxAdapter("db", dbConfig)

	//add the adapter to the worker
	workerRest.SetAdapter(dbAdapter)

	// create a route to the worker
	workerRest.SetRoute("GET", "/", &MainHandler{})

	// append worker to the framework
	radian.AddWorker(workerRest)

	// run the worker
	radian.RunAll()
}
