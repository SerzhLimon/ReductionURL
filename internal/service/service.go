package service

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/model"
	repo "github.com/SerzhLimon/ReductionURL/internal/repository"
)

type UseCase interface {
	SetURL(url string) (string, error)
	GetURL(hash string) (string, error)
	Ping() error
	SetArrayURL(req []model.SetArrayURLRequest) ([]model.SetArrayURLResponse, error)
}

type Service struct {
	repo repo.Repository
}

func NewService(cfg *config.Config, db *sql.DB) (UseCase, error) {
	repo, err := repo.NewStorage(cfg, db)
	if err != nil {
		return nil, err
	}
	return &Service{
		repo: repo,
	}, nil
}

func (s *Service) SetURL(url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", fmt.Errorf("incorrect url")
	}

	hash := sha256.Sum256([]byte(url))
	shortHash := fmt.Sprintf("%x", hash[:8])
	shortHash, err := s.repo.Set(url, shortHash)

	return shortHash, err
}

func (s *Service) GetURL(hash string) (string, error) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return "", fmt.Errorf("incorrect id")
	}
	return s.repo.Get(hash)
}

func (s *Service) Ping() error {
	return s.repo.Ping()
}

func (s *Service) SetArrayURL(req []model.SetArrayURLRequest) ([]model.SetArrayURLResponse, error) {
	for i, item := range req {
		item.OriginalURL = strings.TrimSpace(item.OriginalURL)
		if item.OriginalURL == "" {
			return nil, fmt.Errorf("incorrect url")
		}
		hash := sha256.Sum256([]byte(item.OriginalURL))
		req[i].ShortURL = fmt.Sprintf("%x", hash[:8])
	}
	return s.repo.SetArrayURL(req)
}
