package signaller

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/google/uuid"

	"udpie/internal/model"
	"udpie/internal/model/contract"
	"udpie/pkg/logutils"
)

type pendingResponse struct {
	responseChan chan *contract.WebsocketResponse
	errorChan    chan error
}

type ProducerConnection struct {
	ProducerId      uuid.UUID
	Conn            *websocket.Conn
	mu              sync.Mutex
	pendingRequests map[string]*pendingResponse
	requestsMu      sync.RWMutex
}

type WebsocketService struct {
	mu              sync.RWMutex
	connections     map[uuid.UUID]*ProducerConnection
	producerService contract.SignallerProducerService
}

func NewWebsocketService(producerService contract.SignallerProducerService) *WebsocketService {
	return &WebsocketService{
		connections:     make(map[uuid.UUID]*ProducerConnection),
		producerService: producerService,
	}
}

// registerConnection registers a websocket connection for a producer
func (s *WebsocketService) registerConnection(producerId uuid.UUID, conn *websocket.Conn) error {
	// Verify producer exists
	_, err := s.producerService.GetProducer(producerId)
	if err != nil {
		return errors.New("producer not found")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Close existing connection if any
	if existing, exists := s.connections[producerId]; exists {
		existing.mu.Lock()
		if existing.Conn != nil {
			_ = existing.Conn.Close()
		}
		existing.mu.Unlock()
	}

	s.connections[producerId] = &ProducerConnection{
		ProducerId:      producerId,
		Conn:            conn,
		pendingRequests: make(map[string]*pendingResponse),
	}

	logutils.WithFields(logutils.Fields{
		"producer_id": producerId.String(),
	}).Info("Producer websocket connection registered")

	return nil
}

// removeConnection removes a producer's websocket connection
func (s *WebsocketService) removeConnection(producerId uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, exists := s.connections[producerId]; exists {
		conn.mu.Lock()
		if conn.Conn != nil {
			_ = conn.Conn.Close()
		}
		conn.mu.Unlock()
		delete(s.connections, producerId)

		logutils.WithFields(logutils.Fields{
			"producer_id": producerId.String(),
		}).Info("Producer websocket connection removed")
	}
}

// NotifyProducerAboutTransfer notifies a producer about a new transfer
func (s *WebsocketService) NotifyProducerAboutTransfer(transfer *model.Transfer) error {
	if transfer.FileMeta == nil || transfer.FileMeta.ProducerId == uuid.Nil {
		return errors.New("transfer missing file meta or owner")
	}

	producerId := transfer.FileMeta.ProducerId

	s.mu.RLock()
	conn, exists := s.connections[producerId]
	s.mu.RUnlock()

	if !exists {
		return errors.New("producer not connected")
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.Conn == nil {
		return errors.New("connection closed")
	}

	// Prepare notification message
	notification := map[string]any{
		"type":        "transfer_notification",
		"transfer_id": transfer.Id.String(),
		"file_id":     transfer.FileMeta.Id.String(),
		"file_name":   transfer.FileMeta.Name,
		"file_size":   transfer.FileMeta.Size,
		"consumer": map[string]any{
			"id":          transfer.Consumer.Id.String(),
			"udp_options": transfer.Consumer.UdpOptions,
		},
	}

	message, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	// Send message via websocket
	if err := conn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
		s.removeConnection(producerId)
		return err
	}

	logutils.WithFields(logutils.Fields{
		"producer_id": producerId.String(),
		"transfer_id": transfer.Id.String(),
	}).Info("Notified producer about transfer")

	return nil
}

// MakeClientRequestWithTimeout sends a structured request to producer and waits for structured response.
// This is a blocking operation that will wait for the producer's response or timeout.
//
// Example usage:
//
//	request := contract.WebsocketRequest{
//		ProducerId: producerId,
//		Type: "get_status",
//		Data: map[string]interface{}{
//			"key": "value",
//		},
//	}
//	response, err := service.MakeClientRequestWithTimeout(request, 5*time.Second)
//	if err != nil {
//		// Handle error (timeout, connection closed, etc.)
//	}
//	// Use response.Result or response.Data
//
// Producer response format:
//   - Success: {"request_id": "...", "result": {...}} or {"request_id": "...", "data": {...}}
//   - Error: {"request_id": "...", "error": "error message"}
func (s *WebsocketService) MakeClientRequestWithTimeout(request *contract.WebsocketRequest,
	timeout time.Duration) (*contract.WebsocketResponse, error) {
	s.mu.RLock()
	conn, exists := s.connections[request.ProducerId]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("producer not connected")
	}

	conn.mu.Lock()
	if conn.Conn == nil {
		conn.mu.Unlock()
		return nil, errors.New("connection closed")
	}
	conn.mu.Unlock()

	// Generate request ID if not provided
	if request.RequestID == "" {
		request.RequestID = uuid.New().String()
	}
	requestId := request.RequestID

	// Create pending response channels
	pending := &pendingResponse{
		responseChan: make(chan *contract.WebsocketResponse, 1),
		errorChan:    make(chan error, 1),
	}

	// Register pending request
	conn.requestsMu.Lock()
	conn.pendingRequests[requestId] = pending
	conn.requestsMu.Unlock()

	// Cleanup function
	defer func() {
		conn.requestsMu.Lock()
		delete(conn.pendingRequests, requestId)
		conn.requestsMu.Unlock()
	}()

	// Marshal and send message
	messageBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	conn.mu.Lock()
	if err := conn.Conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
		conn.mu.Unlock()
		return nil, err
	}
	conn.mu.Unlock()

	logutils.WithFields(logutils.Fields{
		"producer_id": request.ProducerId.String(),
		"request_id":  requestId,
		"type":        request.Type,
	}).Debug("Sent message and waiting for response")

	// Wait for response with timeout
	select {
	case response := <-pending.responseChan:
		response.ProducerId = request.ProducerId
		return response, nil
	case err := <-pending.errorChan:
		return nil, err
	case <-time.After(timeout):
		return nil, errors.New("timeout waiting for response")
	}
}

// HandleConnection processes incoming websocket messages from a producer
func (s *WebsocketService) HandleConnection(producerId uuid.UUID, conn *websocket.Conn) error {
	if err := s.registerConnection(producerId, conn); err != nil {
		return err
	}

	// Get producer connection from map
	s.mu.RLock()
	producerConn, exists := s.connections[producerId]
	s.mu.RUnlock()

	if !exists {
		return errors.New("producer connection not found after registration")
	}

	// Handle incoming messages
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			logutils.WithFields(logutils.Fields{
				"producer_id": producerId.String(),
				"error":       err.Error(),
			}).Info("Websocket connection closed")
			break
		}

		// Process message
		if msgType == websocket.TextMessage {
			// Try to parse as JSON to check for request_id
			var messageData map[string]any
			if err := json.Unmarshal(msg, &messageData); err == nil {
				// Check if this is a response to a pending request
				if requestId, ok := messageData["request_id"].(string); ok {
					producerConn.requestsMu.RLock()
					pending, exists := producerConn.pendingRequests[requestId]
					producerConn.requestsMu.RUnlock()

					if exists {
						// Check if this is an error response
						if errorMsg, hasError := messageData["error"].(string); hasError {
							pending.errorChan <- errors.New(errorMsg)
						} else {
							// Parse as WebsocketResponse
							var response contract.WebsocketResponse
							if err := json.Unmarshal(msg, &response); err == nil {
								pending.responseChan <- &response
							} else {
								// Fallback: send error if response parsing failed
								pending.errorChan <- errors.New("failed to parse response")
							}
						}

						logutils.WithFields(logutils.Fields{
							"producer_id": producerId.String(),
							"request_id":  requestId,
						}).Debug("Received response for pending request")
						continue
					}
				}
			}

			logutils.WithFields(logutils.Fields{
				"producer_id": producerId.String(),
				"message":     string(msg),
			}).Debug("Received websocket message from producer")
		}
	}

	// Clean up all pending requests on disconnect
	producerConn.requestsMu.Lock()
	for requestId, pending := range producerConn.pendingRequests {
		// Non-blocking send - channel is buffered with size 1
		select {
		case pending.errorChan <- errors.New("connection closed"):
		default:
			// Channel is full, which shouldn't happen, but handle gracefully
		}
		delete(producerConn.pendingRequests, requestId)
	}
	producerConn.requestsMu.Unlock()

	// Clean up on disconnect
	s.removeConnection(producerId)
	return nil
}
