package service

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net"
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
	GetArrayURL() ([]model.GetArrayURLResponse, error)
	DeleteArrayURL(hashArray []string)
	SetUser(userID int)
	GetStats(subnet string, ip net.IP) (model.GetStatsResponse, error)
}

type Service struct {
	cache *memCache
	repo  repo.Repository
}

func NewService(cfg *config.Config, db *sql.DB) (UseCase, error) {
	cache := newCache()
	pgRepo, err := repo.NewPgStorage(cfg, db)
	if err == nil {
		return &Service{
			cache: cache,
			repo:  pgRepo,
		}, nil
	}
	fileRepo, err := repo.NewFileStorage(cfg)
	if err == nil {
		return &Service{
			cache: cache,
			repo:  fileRepo,
		}, nil
	}
	memRepo, err := repo.NewMemStorage(cfg)
	if err == nil {
		return &Service{
			cache: cache,
			repo:  memRepo,
		}, nil
	}
	return nil, fmt.Errorf("fail to init repo")
}

func (s *Service) SetURL(url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", fmt.Errorf("incorrect url")
	}

	hash := sha256.Sum256([]byte(url))
	shortHash := fmt.Sprintf("%x", hash[:8])

	return s.repo.Set(url, shortHash)
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

func (s *Service) GetArrayURL() ([]model.GetArrayURLResponse, error) {
	return s.repo.GetArrayURL()
}

func (s *Service) DeleteArrayURL(hashArray []string) {
	for _, hash := range hashArray {
		go func(hash string) {
			err := s.repo.Delete(hash)
			if err != nil {
				logrus.Warn(err)
			}
		}(hash)
	}
}

func (s *Service) SetUser(userID int) {
	s.cache.SetUser(userID)
}

func (s *Service) GetStats(subnet string, ip net.IP) (model.GetStatsResponse, error) {

	_, trustedSubnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return model.GetStatsResponse{}, err
	}

	if !trustedSubnet.Contains(ip) {
		return model.GetStatsResponse{}, model.ErrForbiddenIP
	}

	resp, err := s.repo.GetStats()
	if err != nil {
		return model.GetStatsResponse{}, err
	}

	resp.Users = s.cache.GetCountUsers()
	return resp, nil
}
