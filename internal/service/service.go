package service

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	repo "github.com/SerzhLimon/ReductionURL/internal/repository"
)

type UseCase interface {
	SetURL(url string) (string, error)
	GetURL(hash string) (string, error)
}

type Service struct {
	repo repo.Repository
}

func NewService(cfg *config.Config) (UseCase, error) {
	repo, err := repo.NewStorage(cfg)
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
	err := s.repo.Set(url, shortHash)

	return shortHash, err
}

func (s *Service) GetURL(hash string) (string, error) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return "", fmt.Errorf("incorrect id")
	}
	return s.repo.Get(hash)
}
