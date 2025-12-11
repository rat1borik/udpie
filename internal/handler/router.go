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
	wsHandler       *WsHandler
}

func NewRouter(
	producerService contract.SignallerProducerService,
	fileService contract.SignallerFileService,
	transferService contract.SignallerTransferService,
	wsService contract.WebsocketProducerService,
) *Router {
	return &Router{
		producerHandler: NewProducerHandler(producerService),
		fileHandler:     NewFileHandler(fileService, producerService),
		downloadHandler: NewInitDownloadHandler(fileService, producerService, transferService),
		wsHandler:       NewWsHandler(wsService),
	}
}

func (r *Router) SetupRoutes(router *router.Router) {
	apiGroup := router.Group("/api")
	apiGroup.POST("/producers", r.producerHandler.RegisterProducer)
	apiGroup.POST("/files", r.fileHandler.RegisterFile)
	apiGroup.POST("/initDownload", r.downloadHandler.InitDownload)

	// WebSocket endpoint
	router.GET("/ws", r.wsHandler.HandleConnection)

	// Swagger UI
	router.GET("/swagger/{filepath:*}", swagger.WrapHandler())
}
