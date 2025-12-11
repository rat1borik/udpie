package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"udpie/internal/handler"
	"udpie/internal/model"
	"udpie/internal/model/contract"
)

type ConsumerService struct {
	signallerURL string
	client       *http.Client
}

func NewConsumerService(signallerURL string) *ConsumerService {
	return &ConsumerService{
		signallerURL: signallerURL,
		client:       &http.Client{},
	}
}

// InitDownload initiates a file download and returns transfer info
func (s *ConsumerService) InitDownload(fileId uuid.UUID, udpOptions model.UdpOptions) (*contract.InitTransferResult, error) {
	reqBody := &handler.InitDownloadRequest{
		Id:               fileId,
		ClientUdpOptions: udpOptions,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/initDownload", s.signallerURL)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result contract.InitTransferResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
