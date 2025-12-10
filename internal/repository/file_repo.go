package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/model"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	cfg     *config.Config
	mu      sync.RWMutex
	s       []URLRecord
	hasFile bool
}

func NewFileStorage(cfg *config.Config) (Repository, error) {

	s := &FileStorage{
		cfg:     cfg,
		hasFile: cfg.Opts.StorageFile != "",
	}
	if !s.hasFile {
        return nil, fmt.Errorf("fail to init file storage")
    }
	err := s.loadFromFile()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileStorage) loadFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.hasFile {
		return nil
	}

	if _, err := os.Stat(s.cfg.Opts.StorageFile); os.IsNotExist(err) {
		logrus.Warn("File does not exist")
		return nil
	}

	data, err := os.ReadFile(s.cfg.Opts.StorageFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &s.s); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

func (s *FileStorage) Set(url, hash string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := lo.Find(s.s, func(record URLRecord) bool {
		return record.OriginalURL == url
	}); exists {
		existingShortURL, _ := s.Get(hash)
		return existingShortURL, model.ErrURLAlreadyExists
	}

	s.s = append(s.s, URLRecord{
		UUID:        strconv.Itoa(len(s.s)),
		ShortURL:    hash,
		OriginalURL: url,
	})
	err := s.saveToFile()
	return hash, err
}

func (s *FileStorage) Get(hash string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exist := lo.Find(s.s, func(item URLRecord) bool {
		return item.ShortURL == hash
	})
	if !exist {
		return "", fmt.Errorf("%s not found", hash)
	}

	return item.OriginalURL, nil
}

func (s *FileStorage) saveToFile() error {
	dir := filepath.Dir(s.cfg.Opts.StorageFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(s.s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(s.cfg.Opts.StorageFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (s *FileStorage) SetArrayURL(req []model.SetArrayURLRequest) ([]model.SetArrayURLResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp := []model.SetArrayURLResponse{}

	for _, item := range req {
		if existingRecord, exists := lo.Find(s.s, func(record URLRecord) bool {
			return record.OriginalURL == item.OriginalURL
		}); exists {
			resp = append(resp, model.SetArrayURLResponse{
				ID:  item.ID,
				URL: existingRecord.ShortURL,
			})
		} else {
			newRecord := URLRecord{
				UUID:        strconv.Itoa(len(s.s)),
				ShortURL:    item.ShortURL,
				OriginalURL: item.OriginalURL,
			}
			s.s = append(s.s, newRecord)
			resp = append(resp, model.SetArrayURLResponse{
				ID:  item.ID,
				URL: item.ShortURL,
			})
		}
	}

	if err := s.saveToFile(); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *FileStorage) Ping() error {
	return nil
}

func (s *FileStorage) GetArrayURL() ([]model.GetArrayURLResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var res []model.GetArrayURLResponse
	for _, h := range s.s {
		shortURL := s.cfg.Opts.BaseURL + "/" + h.ShortURL
		res = append(res, model.GetArrayURLResponse{Original: h.OriginalURL, Short: shortURL})
	}
	if len(res) == 0 {
		return []model.GetArrayURLResponse{}, fmt.Errorf("not found")
	}

	return res, nil
}

func (s *FileStorage) Delete(hash string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.s = lo.Reject(s.s, func(item URLRecord, _ int) bool {
        return item.ShortURL == hash 
    })

	return s.saveToFile()
}