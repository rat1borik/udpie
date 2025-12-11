package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

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
	reqBody := map[string]interface{}{
		"id":                fileId.String(),
		"client_udp_options": udpOptions,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/initDownload", s.signallerURL)
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
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

