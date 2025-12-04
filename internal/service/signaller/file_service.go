package signaller

import (
	"errors"
	"sync"

	"github.com/google/uuid"

	"udpie/internal/model"
	"udpie/internal/model/contract"
)

type FileService struct {
	mu    sync.RWMutex
	files map[uuid.UUID]*model.FileMeta
}

func NewFileService() *FileService {
	return &FileService{
		files: make(map[uuid.UUID]*model.FileMeta),
	}
}

func (s *FileService) RegisterFile(options contract.RegisterFileOptions) (uuid.UUID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileMeta := model.NewFileMeta(options.Name, options.Size, options.Hash, options.Owner)
	s.files[fileMeta.Id] = fileMeta
	return fileMeta.Id, nil
}

func (s *FileService) GetFileMeta(id uuid.UUID) (*model.FileMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fileMeta, exists := s.files[id]
	if !exists {
		return nil, errors.New("file not found")
	}

	return fileMeta, nil
}
