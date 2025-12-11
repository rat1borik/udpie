package producer

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"udpie/internal/model"
	"udpie/internal/model/contract"
	"udpie/internal/service/common"
)

type WebsocketListener struct {
	producerId      uuid.UUID
	wsURL           string
	conn            *websocket.Conn
	stateService    *StateService
	transferService *TransferService
	stunService     *common.STUNService
}

func NewWebsocketListener(producerId uuid.UUID, signallerURL string,
	stateService *StateService,
	transferService *TransferService,
	stunService *common.STUNService) *WebsocketListener {
	// Convert http:// to ws://
	wsURL := signallerURL
	if len(wsURL) >= 7 && wsURL[:7] == "http://" {
		wsURL = "ws://" + wsURL[7:]
	} else if len(wsURL) >= 8 && wsURL[:8] == "https://" {
		wsURL = "wss://" + wsURL[8:]
	}

	wsURL = fmt.Sprintf("%s/ws?producer_id=%s", wsURL, producerId.String())

	return &WebsocketListener{
		producerId:      producerId,
		wsURL:           wsURL,
		stateService:    stateService,
		transferService: transferService,
		stunService:     stunService,
	}
}

// Listen starts listening for websocket messages
func (w *WebsocketListener) Listen() error {
	fmt.Printf("Connecting to websocket: %s\n", w.wsURL)

	// Connect to websocket
	conn, resp, err := websocket.DefaultDialer.Dial(w.wsURL, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("error connecting to websocket (status %d): %w", resp.StatusCode, err)
		}
		return fmt.Errorf("error connecting to websocket: %w", err)
	}
	w.conn = conn
	defer conn.Close()

	fmt.Printf("Connected to websocket successfully\n")
	fmt.Printf("Listening for messages...\n\n")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Message handling goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Fprintf(os.Stderr, "Websocket error: %v\n", err)
				}
				return
			}

			w.handleMessage(message)
		}
	}()

	// Wait for interrupt or connection close
	select {
	case <-sigChan:
		fmt.Println("\nShutting down...")
		// Close connection gracefully
		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		if err := conn.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
			fmt.Printf("Error sending close message: %v\n", err)
		}
		const closeDelay = 100 * time.Millisecond
		time.Sleep(closeDelay)
	case <-done:
		fmt.Println("Connection closed")
	}

	return nil
}

func (w *WebsocketListener) handleMessage(message []byte) {
	// Try to parse as JSON
	var msgReq contract.WebsocketRequest
	if err := json.Unmarshal(message, &msgReq); err != nil {
		fmt.Printf("Received non-JSON message: %s\n", string(message))
		return
	}

	switch msgReq.Type {
	case "init_transfer":
		j, err := json.Marshal(msgReq.Data)
		if err != nil {
			w.sendErrorResponse(msgReq.RequestID, "failed to parse request data")
			return
		}
		var d model.ProducerInitTransferRequestData
		if err := json.Unmarshal(j, &d); err != nil {
			w.sendErrorResponse(msgReq.RequestID, "failed to parse request data")
			return
		}
		w.handleInitTransferRequest(msgReq.RequestID, &d)
	default:
		fmt.Printf("Unknown request type: %s\n", msgReq.Type)
		// Send error response
		w.sendErrorResponse(msgReq.RequestID, fmt.Sprintf("unknown request type: %s", msgReq.Type))
	}

	fmt.Printf("Received message: %s\n", string(message))
}

func (w *WebsocketListener) handleInitTransferRequest(requestId string, requestData *model.ProducerInitTransferRequestData) {
	// Check if file exists in state
	fileInfo, exists := w.stateService.GetFile(requestData.FileId)
	if !exists {
		w.sendRejectResponse(requestId, "file not found")
		return
	}

	// Check if file exists on disk
	if _, err := os.Stat(fileInfo.FilePath); err != nil {
		w.sendRejectResponse(requestId, fmt.Sprintf("file does not exist: %v", err))
		return
	}

	// Reevaluate udp options
	extAddr, err := w.stunService.Query()
	if err != nil {
		w.sendRejectResponse(requestId, fmt.Sprintf("failed to reevaluate udp options: %v", err))
		return
	}

	udpOptions := model.UdpOptions{}

	udpAddr := extAddr.(*net.UDPAddr)
	udpOptions.ExternalIp = udpAddr.IP.String()
	udpOptions.ExternalPort = udpAddr.Port

	// Accept transfer and send response
	responseData := model.ProducerInitTransferResponseData{
		Status:             model.RequestTransferStatusAccepted,
		ProducerUdpOptions: udpOptions,
	}

	response := contract.WebsocketResponse{
		ProducerId: w.producerId,
		RequestID:  requestId,
		Data:       responseData,
	}

	if writeErr := w.conn.WriteJSON(response); writeErr != nil {
		fmt.Printf("Error sending response: %v\n", writeErr)
		return
	}

	fmt.Printf("Accepted transfer request\n")
	fmt.Printf("  File ID: %s\n", requestData.FileId.String())
	fmt.Printf("  File: %s\n", fileInfo.FilePath)
	fmt.Printf("  Block Size: %d\n", requestData.BlockSize)
	fmt.Printf("  Total Blocks: %d\n", requestData.BlocksCount)
	fmt.Printf("  Consumer: %s:%d\n", requestData.ConsumerUdpOptions.ExternalIp, requestData.ConsumerUdpOptions.ExternalPort)

	// Start file transfer
	consumerAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d",
		requestData.ConsumerUdpOptions.ExternalIp,
		requestData.ConsumerUdpOptions.ExternalPort))
	if err != nil {
		fmt.Printf("Error resolving consumer address: %v\n", err)
		return
	}

	// Generate transfer ID (or use one from request if provided)
	transferId := uuid.New()

	if err := w.transferService.StartTransfer(
		transferId,
		requestData.FileId,
		requestData.BlockSize,
		requestData.BlocksCount,
		consumerAddr,
	); err != nil {
		fmt.Printf("Error starting transfer: %v\n", err)
		return
	}

	fmt.Printf("Transfer started: %s\n", transferId.String())
}

func (w *WebsocketListener) sendErrorResponse(requestId, errorMsg string) {
	response := map[string]any{
		"request_id":  requestId,
		"producer_id": w.producerId.String(),
		"error":       errorMsg,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("Error marshaling error response: %v\n", err)
		return
	}

	if err := w.conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
		fmt.Printf("Error sending error response: %v\n", err)
	}
}

func (w *WebsocketListener) sendRejectResponse(requestId, reason string) {
	responseData := model.ProducerInitTransferResponseData{
		Status: model.RequestTransferStatusRejected,
	}

	response := map[string]any{
		"request_id":  requestId,
		"producer_id": w.producerId.String(),
		"data":        responseData,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("Error marshaling reject response: %v\n", err)
		return
	}

	if err := w.conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
		fmt.Printf("Error sending reject response: %v\n", err)
	}

	fmt.Printf("Rejected transfer request: %s\n", reason)
}
