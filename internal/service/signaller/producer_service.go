package signaller

import (
	"errors"
	"sync"

	"github.com/google/uuid"

	"udpie/internal/model"
	"udpie/internal/model/contract"
)

type ProducerService struct {
	mu        sync.RWMutex
	producers map[uuid.UUID]*model.Producer
}

func NewProducerService() *ProducerService {
	return &ProducerService{
		producers: make(map[uuid.UUID]*model.Producer),
	}
}

func (s *ProducerService) RegisterProducer(options contract.RegisterProducerOptions) (uuid.UUID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	producer := model.NewProducer(options.UdpOptions)
	id := producer.Id

	s.producers[id] = producer
	return id, nil
}

func (s *ProducerService) GetProducer(id uuid.UUID) (*model.Producer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	producer, exists := s.producers[id]
	if !exists {
		return nil, errors.New("producer not found")
	}

	return producer, nil
}

func (s *ProducerService) UpdateProducer(id uuid.UUID, options contract.RegisterProducerOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	producer, exists := s.producers[id]
	if !exists {
		return errors.New("producer not found")
	}

	producer.UdpOptions = options.UdpOptions
	return nil
}
