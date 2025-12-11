package handler

import (
	"github.com/fasthttp/websocket"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"udpie/internal/model/contract"
	"udpie/pkg/logutils"
)

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WsHandler struct {
	service contract.WebsocketProducerService
}

func NewWsHandler(service contract.WebsocketProducerService) *WsHandler {
	return &WsHandler{
		service: service,
	}
}

// HandleConnection handles websocket connection upgrade and registration
// @Summary      WebSocket connection for producers
// @Description  Establishes a WebSocket connection for a producer to receive transfer notifications
// @Tags         websocket
// @Param        producer_id  query  string  true  "Producer ID"
// @Success      101          "Switching Protocols"
// @Failure      400          {object}  map[string]any  "Invalid producer ID"
// @Failure      500          {object}  map[string]any  "Internal server error"
// @Router       /ws [get]
func (h *WsHandler) HandleConnection(ctx *fasthttp.RequestCtx) {
	// Get producer ID from query parameter
	producerIdStr := string(ctx.QueryArgs().Peek("producer_id"))
	if producerIdStr == "" {
		ErrorWithMessage(ctx, fasthttp.StatusBadRequest, "producer_id query parameter is required")
		return
	}

	producerId, err := uuid.Parse(producerIdStr)
	if err != nil {
		ErrorWithMessage(ctx, fasthttp.StatusBadRequest, "invalid producer_id format")
		return
	}

	// Upgrade to websocket
	err = upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		defer func() {
			if r := recover(); r != nil {
				logutils.WithFields(logutils.Fields{
					"producer_id": producerId.String(),
					"error":       r,
				}).Error("Panic in websocket handler")
			}
		}()

		// Handle the connection
		if err2 := h.service.HandleConnection(producerId, conn); err2 != nil {
			logutils.WithFields(logutils.Fields{
				"producer_id": producerId.String(),
				"error":       err2.Error(),
			}).Error("Error handling websocket connection")
		}
	})

	if err != nil {
		ErrorWithMessage(ctx, fasthttp.StatusBadRequest, "Failed to upgrade to websocket")
		return
	}
}
