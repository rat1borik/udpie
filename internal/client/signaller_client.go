package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"udpie/internal/model"
	"udpie/internal/model/contract"
)

type SignallerClient struct {
	baseURL string
	client  *http.Client
}

func NewSignallerClient(baseURL string) *SignallerClient {
	return &SignallerClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// RegisterProducer registers a producer and returns the producer ID
func (c *SignallerClient) RegisterProducer(udpOptions model.UdpOptions) (uuid.UUID, error) {
	reqBody := contract.RegisterProducerOptions{
		UdpOptions: udpOptions,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/producers", c.baseURL)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return uuid.Nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return uuid.Nil, fmt.Errorf("failed to decode response: %w", decodeErr)
	}

	producerId, err := uuid.Parse(result.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid producer ID format: %w", err)
	}

	return producerId, nil
}

// RegisterFile registers a file and returns the file ID
func (c *SignallerClient) RegisterFile(name string, size uint64, producerId uuid.UUID) (uuid.UUID, error) {
	reqBody := map[string]any{
		"name":        name,
		"size":        size,
		"producer_id": producerId.String(),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/files", c.baseURL)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return uuid.Nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
		return uuid.Nil, fmt.Errorf("failed to decode response: %w", decodeErr)
	}

	fileId, err := uuid.Parse(result.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid file ID format: %w", err)
	}

	return fileId, nil
}
