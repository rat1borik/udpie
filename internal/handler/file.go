//nolint:dupl
package handler

import (
	"encoding/json"

	"github.com/valyala/fasthttp"

	"udpie/internal/model/contract"
)

type FileHandler struct {
	service contract.SignallerFileService
}

func NewFileHandler(service contract.SignallerFileService) *FileHandler {
	return &FileHandler{
		service: service,
	}
}

func (h *FileHandler) RegisterFile(ctx *fasthttp.RequestCtx) {
	if !ctx.IsPost() {
		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
		return
	}

	var options contract.RegisterFileOptions
	if err := json.Unmarshal(ctx.PostBody(), &options); err != nil {
		ctx.Error("Invalid request body", fasthttp.StatusBadRequest)
		return
	}

	id, err := h.service.RegisterFile(options)
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
