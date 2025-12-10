package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/SerzhLimon/ReductionURL/internal/model"
	"github.com/sirupsen/logrus"
)

type PgStorage struct {
	cfg *config.Config
	db  *sql.DB
}

func NewPgStorage(cfg *config.Config, db *sql.DB) (Repository, error) {

	s := &PgStorage{
		cfg: cfg,
		db:  db,
	}

	if db == nil {
		return nil, fmt.Errorf("db not init")
	}

	return s, nil
}

func (s *PgStorage) Ping() error {
	if s.db == nil {
		return fmt.Errorf("db is not init")
	}
	return s.db.Ping()
}

func (s *PgStorage) Set(url, hash string) (string, error) {
	result, err := s.db.Exec(querySetURL, url, hash)

	if err != nil {
		return "", err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", err
	}

	if rowsAffected == 0 {
		return hash, model.ErrURLAlreadyExists
	}

	return hash, nil
}

func (s *PgStorage) Get(hash string) (string, error) {
	var originalURL string
	var isDeleted bool
	err := s.db.QueryRow(queryGetURL, hash).Scan(&originalURL, &isDeleted)

	if err != nil {
		if isDeleted {
			logrus.Error("PgStorage1")
			return "", model.ErrDeletedURL
		}
		if errors.Is(err, sql.ErrNoRows) {
			logrus.Error("PgStorage2")
			return "", fmt.Errorf("URL not found")
		}
		logrus.Error("PgStorage3")
		return "", fmt.Errorf("database error: %w", err)
	}

	return originalURL, nil
}

func (s *PgStorage) SetArrayURL(req []model.SetArrayURLRequest) ([]model.SetArrayURLResponse, error) {
	tx, err := s.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	resp := []model.SetArrayURLResponse{}
	for _, item := range req {
		_, err := tx.Exec(querySetURL, item.OriginalURL, item.ShortURL)
		if err != nil {
			return nil, err
		}
		resp = append(resp, model.SetArrayURLResponse{
			ID:  item.ID,
			URL: s.cfg.Opts.BaseURL + "/" + item.ShortURL,
		})
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *PgStorage) GetArrayURL() ([]model.GetArrayURLResponse, error) {
	var res []model.GetArrayURLResponse

	rows, err := s.db.Query(queryGetArrayURL)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// пробегаем по всем записям
	for rows.Next() {
		var v model.GetArrayURLResponse
		err = rows.Scan(&v.Original, &v.Short)
		if err != nil {
			return nil, err
		}
		shortURL := s.cfg.Opts.BaseURL + "/" + v.Short
		v.Short = shortURL

		res = append(res, v)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *PgStorage) Delete(hash string) error {
	result, err := s.db.Exec(queryDeleteURL, hash)
	if err != nil {
		return fmt.Errorf("failed delete %s; %w", hash, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed delete %s; %w", hash, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("failed delete %s; rows affected == 0", hash)
	}

	return nil
}
