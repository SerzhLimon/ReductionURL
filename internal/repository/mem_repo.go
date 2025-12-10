package repository

import (
	"fmt"
	"sync"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/model"
)

type DataURL struct {
	OriginalURL string
	IsDeleted   bool
}

type MemStorage struct {
	cfg         *config.Config
	memoryCache map[string]DataURL
	mu          sync.RWMutex
}

func NewMemStorage(cfg *config.Config) (Repository, error) {

	s := &MemStorage{
		cfg:         cfg,
		memoryCache: make(map[string]DataURL),
	}

	return s, nil
}

func (s *MemStorage) Get(hash string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exist := s.memoryCache[hash]
	if !exist {
		return "", fmt.Errorf("%s not found", hash)
	}
	if data.IsDeleted {
		return "", model.ErrDeletedURL
	}
	return data.OriginalURL, nil
}

func (s *MemStorage) Set(url, hash string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exist := s.memoryCache[hash]; exist {
		existingShortURL, _ := s.Get(hash)
		return existingShortURL, model.ErrURLAlreadyExists
	}

	s.memoryCache[hash] = DataURL{OriginalURL: url}
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

func (s *MemStorage) GetArrayURL() ([]model.GetArrayURLResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var res []model.GetArrayURLResponse
	for key, val := range s.memoryCache {
		shortURL := s.cfg.Opts.BaseURL + "/" + key
		res = append(res, model.GetArrayURLResponse{Original: val.OriginalURL, Short: shortURL})
	}

	if len(res) == 0 {
		return []model.GetArrayURLResponse{}, fmt.Errorf("not found")
	}

	return res, nil
}

func (s *MemStorage) Delete(hash string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	delete(s.memoryCache, hash)
	return nil
}
