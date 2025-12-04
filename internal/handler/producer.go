//nolint:dupl
package handler

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	"udpie/internal/model/contract"
)

type ProducerHandler struct {
	service contract.SignallerProducerService
}

func NewProducerHandler(service contract.SignallerProducerService) *ProducerHandler {
	return &ProducerHandler{
		service: service,
	}
}

// RegisterProducer registers a new producer
// @Summary      Register a new producer
// @Description  Register a new producer with UDP options for file transfer
// @Tags         producers
// @Accept       json
// @Produce      json
// @Param        request  body      contract.RegisterProducerOptions  true  "Producer registration options"
// @Success      200      {object}  map[string]any  "Success response with producer ID"
// @Failure      400      {object}  map[string]any  "Invalid request body"
// @Failure      500      {object}  map[string]any  "Internal server error"
// @Router       /producers [post]
func (h *ProducerHandler) RegisterProducer(ctx *fasthttp.RequestCtx) {
	var options contract.RegisterProducerOptions
	if err := json.Unmarshal(ctx.PostBody(), &options); err != nil {
		ctx.Error("Invalid request body", fasthttp.StatusBadRequest)
		return
	}

	id, err := h.service.RegisterProducer(options)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	Success(ctx, map[string]string{"id": id.String()})
}
