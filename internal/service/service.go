package service

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/model"
	repo "github.com/SerzhLimon/ReductionURL/internal/repository"
	"github.com/sirupsen/logrus"
)

type UseCase interface {
	SetURL(url string) (string, error)
	GetURL(hash string) (string, error)
	Ping() error
	SetArrayURL(req []model.SetArrayURLRequest) ([]model.SetArrayURLResponse, error)
}

type Service struct {
	pgRepo   repo.Repository
	fileRepo repo.Repository
	memRepo  repo.Repository
}

func NewService(cfg *config.Config, db *sql.DB) (UseCase, error) {
	pgRepo, err := repo.NewPgStorage(cfg, db)
	if err != nil {
		logrus.Warn(err)
	}
	fileRepo, err := repo.NewFileStorage(cfg)
	if err != nil {
		logrus.Warn(err)
	}
	memRepo, err := repo.NewMemStorage(cfg)
	if err != nil {
		logrus.Warn(err)
	}
	fatal := memRepo == nil && fileRepo == nil && pgRepo == nil
	if fatal {
		return nil, fmt.Errorf("fail to init repository")
	}
	return &Service{
		pgRepo:   pgRepo,
		fileRepo: fileRepo,
		memRepo:  memRepo,
	}, nil
}

func (s *Service) SetURL(url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", fmt.Errorf("incorrect url")
	}

	hash := sha256.Sum256([]byte(url))
	shortHash := fmt.Sprintf("%x", hash[:8])

	switch {
	case s.pgRepo != nil:
		return s.pgRepo.Set(url, shortHash)
	case s.fileRepo != nil:
		return s.fileRepo.Set(url, shortHash)
	}

	return s.memRepo.Set(url, shortHash)
}

func (s *Service) GetURL(hash string) (string, error) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return "", fmt.Errorf("incorrect id")
	}
	switch {
	case s.pgRepo != nil:
		return s.pgRepo.Get(hash)
	case s.fileRepo != nil:
		return s.fileRepo.Get(hash)
	}
	return s.memRepo.Get(hash)
}

func (s *Service) Ping() error {
	if s.pgRepo != nil {
		return s.pgRepo.Ping()
	}
	return nil
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
	switch {
	case s.pgRepo != nil:
		return s.pgRepo.SetArrayURL(req)
	case s.fileRepo != nil:
		return s.fileRepo.SetArrayURL(req)
	}
	return s.memRepo.SetArrayURL(req)
}
