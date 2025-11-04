package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/SerzhLimon/ReductionURL/internal/config"
)

type Repository interface {
	Get(hash string) (string, error)
	Set(url, hash string) error
}

type FileStorage struct {
	cfg *config.Config
	s   map[string]string
	mu  sync.RWMutex
}

func NewStorage(cfg *config.Config) (Repository, error) {
	storage := make(map[string]string, 50)

	fs := &FileStorage{
		s:   storage,
		cfg: cfg,
	}

	err := fs.loadFromFile()
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func (fs *FileStorage) loadFromFile() error {
	absPath, _ := filepath.Abs(fs.cfg.Opts.StorageFile)
	fs.mu.Lock()
	defer fs.mu.Unlock()

	file, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		return nil
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&fs.s); err != nil {
		return err
	}

	return nil
}

func (fs *FileStorage) Get(hash string) (string, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	url, exist := fs.s[hash]
	if !exist {
		return "", fmt.Errorf("%s not found", hash)
	}
	return url, nil
}

func (fs *FileStorage) Set(url, hash string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.s[hash] = url
	return fs.saveToFile()
}

func (fs *FileStorage) saveToFile() error {
	absPath, _ := filepath.Abs(fs.cfg.Opts.StorageFile)
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(fs.cfg.Opts.StorageFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(fs.s); err != nil {
		return err
	}
	
	return nil
}
