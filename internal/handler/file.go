//nolint:dupl
package handler

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"udpie/internal/model/contract"
)

type RegisterFileRequest struct {
	Name       string    `json:"name"`
	Size       uint64    `json:"size"`
	ProducerId uuid.UUID `json:"producer_id"`
}

type FileHandler struct {
	service         contract.SignallerFileService
	producerService contract.SignallerProducerService
}

func NewFileHandler(service contract.SignallerFileService, producerService contract.SignallerProducerService) *FileHandler {
	return &FileHandler{
		service:         service,
		producerService: producerService,
	}
}

// RegisterFile registers a new file
// @Summary      Register a new file
// @Description  Register a new file with metadata and associate it with a producer
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterFileRequest  true  "File registration request"
// @Success      200      {object}  map[string]any  "Success response with file ID"
// @Failure      400      {object}  map[string]any  "Invalid request body"
// @Failure      500      {object}  map[string]any  "Internal server error"
// @Router       /files [post]
func (h *FileHandler) RegisterFile(ctx *fasthttp.RequestCtx) {
	var request RegisterFileRequest
	if err := json.Unmarshal(ctx.PostBody(), &request); err != nil {
		ctx.Error("Invalid request body", fasthttp.StatusBadRequest)
		return
	}

	_, err := h.producerService.GetProducer(request.ProducerId)
	if err != nil {
		ctx.Error("Producer not found", fasthttp.StatusBadRequest)
		return
	}

	id, err := h.service.RegisterFile(contract.RegisterFileOptions{
		Name:       request.Name,
		Size:       request.Size,
		ProducerId: request.ProducerId,
	})
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	Success(ctx, map[string]string{"id": id.String()})
}
