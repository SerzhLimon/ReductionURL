package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Opts *Options
}

type Options struct {
	Addr         string `env:"SERVER_ADDRESS"`
	BaseURL      string `env:"BASE_URL"`
	StorageFile  string `env:"FILE_STORAGE_PATH"`
	DataBaseHost string `env:"DATABASE_DSN"`
}

func NewConfig() (*Config, error) {
	opts, err := newOpts()
	if err != nil {
		return nil, err
	}
	return &Config{
		Opts: opts,
	}, nil
}

func newOpts() (*Options, error) {
	envAddr := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envStorageFile := os.Getenv("FILE_STORAGE_PATH")
	envPsqlHost := os.Getenv("DATABASE_DSN")

	storageFileValue := "storage.json"
	if envStorageFile != "" {
		storageFileValue = envStorageFile
	}

	dataBaseHost := "localhost:5432"
	if envPsqlHost != "" {
		dataBaseHost = envPsqlHost
	}

	if envAddr != "" && envBaseURL != "" {
		if _, err := url.Parse("http://" + envAddr); err == nil {
			if _, err := url.Parse(envBaseURL); err == nil {
				return &Options{
					Addr:         envAddr,
					BaseURL:      envBaseURL,
					StorageFile:  storageFileValue,
					DataBaseHost: dataBaseHost,
				}, nil
			}
		}
	}

	var addr = flag.String("a", "localhost:8080", "server host")
	var baseURL = flag.String("b", "localhost:8080", "value before short URL")
	var storageFile = flag.String("f", "storage.json", "file for save data")
	var psqlHost = flag.String("d", "localhost:5432", "psql host")
	flag.Parse()

	addrValue := *addr
	if envAddr != "" {
		addrValue = envAddr
	}

	baseURLValue := *baseURL
	if envBaseURL != "" {
		baseURLValue = envBaseURL
	}

	if _, err := url.Parse("https://" + addrValue); err != nil {
		return nil, fmt.Errorf("incorrect parametr `-a` %s", addrValue)
	}

	if _, err := url.Parse("https://" + baseURLValue); err != nil {
		return nil, fmt.Errorf("incorrect parametr `-b` %s", baseURLValue)
	}

	if !strings.HasPrefix(baseURLValue, "http://") && !strings.HasPrefix(baseURLValue, "https://") {
		baseURLValue = "http://" + baseURLValue
	}
	baseURLValue = strings.TrimSuffix(baseURLValue, "/")

	if envStorageFile != "" {
		storageFileValue = envStorageFile
	} else {
		storageFileValue = *storageFile
	}

	if envPsqlHost != "" {
		dataBaseHost = envPsqlHost
	} else {
		dataBaseHost = *psqlHost
	}

	return &Options{
		Addr:         addrValue,
		BaseURL:      baseURLValue,
		StorageFile:  storageFileValue,
		DataBaseHost: dataBaseHost,
	}, nil
}
