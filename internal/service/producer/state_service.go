package producer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const (
	defaultStateFile = ".udpie-producer-state.json"
)

type ProducerState struct {
	ProducerId uuid.UUID           `json:"producer_id"`
	Files      map[string]FileInfo `json:"files"` // fileId -> FileInfo
}

type FileInfo struct {
	FileId   uuid.UUID `json:"file_id"`
	Name     string    `json:"name"`
	Size     uint64    `json:"size"`
	FilePath string    `json:"file_path"`
}

type StateService struct {
	stateFile string
	state     *ProducerState
}

func NewStateService(stateFile string) *StateService {
	if stateFile == "" {
		stateFile = defaultStateFile
	}

	return &StateService{
		stateFile: stateFile,
		state: &ProducerState{
			Files: make(map[string]FileInfo),
		},
	}
}

// Load loads state from file
func (s *StateService) Load() error {
	data, err := os.ReadFile(s.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use empty state
			return nil
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if err := json.Unmarshal(data, s.state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	if s.state.Files == nil {
		s.state.Files = make(map[string]FileInfo)
	}

	return nil
}

// Save saves state to file
func (s *StateService) Save() error {
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.stateFile)
	if dir != "." && dir != "" {
		const dirPerm = 0755
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	const filePerm = 0600
	if err := os.WriteFile(s.stateFile, data, filePerm); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// GetProducerId returns the saved producer ID
func (s *StateService) GetProducerId() (uuid.UUID, bool) {
	if s.state.ProducerId == uuid.Nil {
		return uuid.Nil, false
	}
	return s.state.ProducerId, true
}

// SetProducerId sets and saves the producer ID
func (s *StateService) SetProducerId(producerId uuid.UUID) error {
	s.state.ProducerId = producerId
	return s.Save()
}

// AddFile adds a file to the state
func (s *StateService) AddFile(fileId uuid.UUID, name string, size uint64, filePath string) error {
	s.state.Files[fileId.String()] = FileInfo{
		FileId:   fileId,
		Name:     name,
		Size:     size,
		FilePath: filePath,
	}
	return s.Save()
}

// GetFile returns file info by file ID
func (s *StateService) GetFile(fileId uuid.UUID) (FileInfo, bool) {
	info, exists := s.state.Files[fileId.String()]
	return info, exists
}

// GetAllFiles returns all registered files
func (s *StateService) GetAllFiles() map[string]FileInfo {
	return s.state.Files
}
