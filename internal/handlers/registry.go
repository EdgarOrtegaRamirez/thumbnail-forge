package handlers

import (
	"sync"

	"github.com/EdgarOrtegaRamirez/thumbnail-forge/internal/models"
)

var (
	registryMutex sync.RWMutex
	handlersList  []models.Handler
)

// Register adds a handler to the global registry
func Register(h models.Handler) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	handlersList = append(handlersList, h)
}

// GetHandler finds the first registered handler that can process the given file info.
// Returns nil if no handler is found.
func GetHandler(info *models.FileInfo) models.Handler {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	for _, h := range handlersList {
		if h.CanHandle(info) {
			return h
		}
	}
	return nil
}

func init() {
	Register(&ImageHandler{})
	Register(&CodeHandler{})
	Register(&PDFHandler{})
	Register(&VideoHandler{})
	Register(&AudioHandler{})
	Register(&OfficeHandler{})
	Register(&ArchiveHandler{})
	Register(&DiskImageHandler{})
}
