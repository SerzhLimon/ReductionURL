package repository

import (
	"fmt"
	"sync"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/model"
)

type MemStorage struct {
	cfg         *config.Config
	memoryCache map[string]string
	mu          sync.RWMutex
}

func NewMemStorage(cfg *config.Config) (Repository, error) {

	s := &MemStorage{
		cfg:         cfg,
		memoryCache: make(map[string]string),
	}

	return s, nil
}

func (s *MemStorage) Get(hash string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, exist := s.memoryCache[hash]
	if !exist {
		return "", fmt.Errorf("%s not found", hash)
	}
	return url, nil
}

func (s *MemStorage) Set(url, hash string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exist := s.memoryCache[hash]; exist {
		existingShortURL, _ := s.Get(hash)
		return existingShortURL, model.ErrURLAlreadyExists
	}

	s.memoryCache[hash] = url
	return hash, nil
}

func (s *MemStorage) SetArrayURL(req []model.SetArrayURLRequest) ([]model.SetArrayURLResponse, error) {
	resp := []model.SetArrayURLResponse{}
	for _, item := range req {
		s.Set(item.OriginalURL, item.ShortURL)
		resp = append(resp, model.SetArrayURLResponse{
			ID:  item.ID,
			URL: item.ShortURL,
		})
	}
	return resp, nil
}

func (s *MemStorage) Ping() error {
	return nil
}