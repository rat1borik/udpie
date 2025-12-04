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

func (h *ProducerHandler) RegisterProducer(ctx *fasthttp.RequestCtx) {
	if !ctx.IsPost() {
		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
		return
	}

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

	ctx.SetContentType("application/json")
	response := map[string]string{"id": id.String()}
	if body, err := json.Marshal(response); err == nil {
		ctx.SetBody(body)
	} else {
		ctx.Error("Failed to encode response", fasthttp.StatusInternalServerError)
	}
}
