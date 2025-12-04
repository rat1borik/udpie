package handler

import (
	"github.com/fasthttp/router"

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
		fileHandler:     NewFileHandler(fileService),
	}
}

func (r *Router) SetupRoutes(router *router.Router) {
	router.POST("/api/producers", r.producerHandler.RegisterProducer)
	router.POST("/api/files", r.fileHandler.RegisterFile)
}
