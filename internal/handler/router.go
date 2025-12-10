package handler

import (
	"github.com/fasthttp/router"
	swagger "github.com/swaggo/fasthttp-swagger"

	"udpie/internal/model/contract"
)

type Router struct {
	producerHandler *ProducerHandler
	fileHandler     *FileHandler
	downloadHandler *InitDownloadHandler
}

func NewRouter(
	producerService contract.SignallerProducerService,
	fileService contract.SignallerFileService,
	transferService contract.SignallerTransferService,
) *Router {
	return &Router{
		producerHandler: NewProducerHandler(producerService),
		fileHandler:     NewFileHandler(fileService, producerService),
		downloadHandler: NewInitDownloadHandler(fileService, producerService, transferService),
	}
}

func (r *Router) SetupRoutes(router *router.Router) {
	apiGroup := router.Group("/api")
	apiGroup.POST("/producers", r.producerHandler.RegisterProducer)
	apiGroup.POST("/files", r.fileHandler.RegisterFile)
	apiGroup.POST("/initDownload", r.downloadHandler.InitDownload)

	// Swagger UI
	router.GET("/swagger/{filepath:*}", swagger.WrapHandler())
}
