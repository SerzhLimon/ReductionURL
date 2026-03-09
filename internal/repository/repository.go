package repository

import "github.com/SerzhLimon/ReductionURL/internal/model"

type Repository interface {
	Get(hash string) (string, error)
	Set(url, hash string) (string, error)
	Ping() error
	SetArrayURL(req []model.SetArrayURLRequest) ([]model.SetArrayURLResponse, error)
	GetArrayURL() ([]model.GetArrayURLResponse, error)
	Delete(hash string) error
}
