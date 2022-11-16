package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/radianteam/framework/worker"
)

type RestServiceHandlerInterface interface {
	worker.BaseHandlerInterface

	SetGinContext(*gin.Context)
}

type RestServiceHandler struct {
	worker.BaseHandler

	GinContext *gin.Context
}

func (h *RestServiceHandler) SetGinContext(c *gin.Context) {
	h.GinContext = c
}
