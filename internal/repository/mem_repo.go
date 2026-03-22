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
	memoryCache sync.Map
}

func NewMemStorage(cfg *config.Config) (Repository, error) {
	s := &MemStorage{
		cfg:         cfg,
		memoryCache: sync.Map{},
	}

	return s, nil
}

func (s *MemStorage) Get(hash string) (string, error) {
	actual, exist := s.memoryCache.Load(hash)
	if !exist {
		return "", fmt.Errorf("%s not found", hash)
	}
	data, _ := actual.(DataURL)
	if data.IsDeleted {
		return "", model.ErrDeletedURL
	}
	return data.OriginalURL, nil
}

func (s *MemStorage) Set(url, hash string) (string, error) {
	
	if data, exist := s.memoryCache.Load(hash); exist {
		existingShortURL, _  := data.(DataURL)
		return existingShortURL.OriginalURL, model.ErrURLAlreadyExists
	}

	s.memoryCache.Store(hash, DataURL{OriginalURL: url})
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
	var res []model.GetArrayURLResponse
	
	s.memoryCache.Range(func(key, value interface{}) bool {
		strKey := key.(string)
		data := value.(DataURL)
		
		shortURL := s.cfg.Opts.BaseURL + "/" + strKey
		res = append(res, model.GetArrayURLResponse{
			Original: data.OriginalURL, 
			Short:    shortURL,
		})
		
		return true
	})

	if len(res) == 0 {
		return []model.GetArrayURLResponse{}, fmt.Errorf("not found")
	}

	return res, nil
}

func (s *MemStorage) Delete(hash string) error {
	s.memoryCache.Delete(hash)
	return nil
}

func (s *MemStorage) GetStats() (model.GetStatsResponse, error) {
	var countHashURL int
	s.memoryCache.Range(func(_, value interface{}) bool {
		data := value.(DataURL)
		if !data.IsDeleted {
			countHashURL++
		}
		return true
	})
	return model.GetStatsResponse{URLs: countHashURL}, nil
}