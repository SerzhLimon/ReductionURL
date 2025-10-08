package service

import (
	"crypto/sha256"
	"fmt"
	"strings"

	repo "github.com/SerzhLimon/ReductionURL/internal/repository"
)

type Service struct {
	repo *repo.MemStorage
}

func NewService() *Service {
	return &Service{
		repo: repo.NewStrorage(),
	}
}

func (s *Service) SetURL(url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", fmt.Errorf("incorrect url")
	}

	hash := sha256.Sum256([]byte(url))
	shortHash := fmt.Sprintf("%x", hash[:8])
	s.repo.Set(url, shortHash)

	return shortHash, nil
}

func (s *Service) GetURL(hash string) (string, error) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return "", fmt.Errorf("incorrect id")
	}
	return s.repo.Get(hash)
}
