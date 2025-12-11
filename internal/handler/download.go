//nolint:dupl
package handler

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"udpie/internal/model"
	"udpie/internal/model/contract"
)

type InitDownloadRequest struct {
	Id               uuid.UUID        `json:"id"`
	ClientUdpOptions model.UdpOptions `json:"client_udp_options"`
}

type InitDownloadHandler struct {
	service         contract.SignallerFileService
	producerService contract.SignallerProducerService
	transferService contract.SignallerTransferService
}

func NewInitDownloadHandler(service contract.SignallerFileService,
	producerService contract.SignallerProducerService,
	transferService contract.SignallerTransferService) *InitDownloadHandler {
	return &InitDownloadHandler{
		service:         service,
		producerService: producerService,
		transferService: transferService,
	}
}

// InitDownload inits download of a file
// @Summary      Init download
// @Description
// @Tags         download
// @Accept       json
// @Produce      json
// @Param        request  body      InitDownloadRequest  true  "Init download request"
// @Success      200      {object}  map[string]any  "Success response with transfer ID"
// @Failure      400      {object}  map[string]any  "Invalid request body"
// @Failure      500      {object}  map[string]any  "Internal server error"
// @Router       /initDownload [post]
func (h *InitDownloadHandler) InitDownload(ctx *fasthttp.RequestCtx) {
	var request InitDownloadRequest
	if err := json.Unmarshal(ctx.PostBody(), &request); err != nil {
		ctx.Error("Invalid request body", fasthttp.StatusBadRequest)
		return
	}

	transfer, err := h.transferService.InitTransfer(contract.InitTransferOptions{
		FileId:             request.Id,
		ConsumerUdpOptions: request.ClientUdpOptions,
	})
	if err != nil {
		ctx.Error(fmt.Sprintf("Failed to init download: %v", err), fasthttp.StatusInternalServerError)
		return
	}

	Success(ctx, transfer)
}
