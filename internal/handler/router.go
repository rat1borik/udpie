package handler

import (
	"github.com/fasthttp/router"
	swagger "github.com/swaggo/fasthttp-swagger"

	"udpie/internal/model/contract"
)

type Router struct {
	producerHandler *ProducerHandler
	fileHandler     *FileHandler
}

func NewRouter(
	producerService contract.SignallerProducerService,
	fileService contract.SignallerFileService,
) *Router {
	return &Router{
		producerHandler: NewProducerHandler(producerService),
		fileHandler:     NewFileHandler(fileService, producerService),
	}
}

func (r *Router) SetupRoutes(router *router.Router) {
	apiGroup := router.Group("/api")
	apiGroup.POST("/producers", r.producerHandler.RegisterProducer)
	apiGroup.POST("/files", r.fileHandler.RegisterFile)

	// Swagger UI
	router.GET("/swagger/{filepath:*}", swagger.WrapHandler())
}
