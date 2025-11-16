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

	opts, ok := parseEnv()
	if ok {
		return opts, nil
	}
	var addr = flag.String("a", "localhost:8080", "server host")
	var baseURL = flag.String("b", "localhost:8080", "value before short URL")
	var storageFile = flag.String("f", "storage.json", "file for save data")
	var psqlHost = flag.String("d", "", "psql data")
	flag.Parse()

	
	if opts.Addr == "" {
		opts.Addr = *addr
	}

	// baseURLValue := *baseURL
	if opts.BaseURL == "" {
		opts.BaseURL = *baseURL
	}
	if opts.StorageFile == "storage.json" {
		opts.StorageFile = *storageFile
	}
	// dataBaseHost := *psqlHost
	if opts.DataBaseHost == "" {
		opts.DataBaseHost = *psqlHost
	}
	if _, err := url.Parse("https://" + opts.Addr); err != nil {
		return nil, fmt.Errorf("incorrect parametr `-a` %s", opts.Addr)
	}

	if _, err := url.Parse("https://" + opts.BaseURL); err != nil {
		return nil, fmt.Errorf("incorrect parametr `-b` %s", opts.BaseURL)
	}

	if _, err := url.Parse("https://" + opts.DataBaseHost); err != nil {
		return nil, fmt.Errorf("incorrect parametr `-d` %s", opts.DataBaseHost)
	}

	if !strings.HasPrefix(opts.BaseURL, "http://") && !strings.HasPrefix(opts.BaseURL, "https://") {
		opts.BaseURL = "http://" + opts.BaseURL
	}
	opts.BaseURL = strings.TrimSuffix(opts.BaseURL, "/")

	// if envStorageFile != "" {
	// 	storageFileValue = envStorageFile
	// } else {
	// 	storageFileValue = *storageFile
	// }

	// if envPsqlHost != "" {
	// 	dataBaseHost = envPsqlHost
	// } else {
	// 	dataBaseHost = *psqlHost
	// }

	return opts, nil
}

func parseEnv() (*Options, bool) {
	envAddr := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envStorageFile := os.Getenv("FILE_STORAGE_PATH")
	envPsqlDsn := os.Getenv("DATABASE_DSN")

	opts := &Options{
		StorageFile: "storage.json",
	}
	if envStorageFile != "" {
		opts.StorageFile = envStorageFile
	}
	if _, err := url.Parse("http://" + envAddr); err == nil {
		opts.Addr = envAddr
	}
	if _, err := url.Parse(envBaseURL); err == nil { 
		opts.BaseURL = envBaseURL
	}
	if _, err := url.Parse(envPsqlDsn); err == nil { 
		opts.DataBaseHost = envPsqlDsn
	}

	sucessAll := opts.Addr != "" && opts.BaseURL != "" && opts.DataBaseHost != ""
	return opts, sucessAll
	// if envAddr != "" && envBaseURL != "" {
	// 	if _, err := url.Parse("http://" + envAddr); err == nil {
	// 		if _, err := url.Parse(envBaseURL); err == nil { 
	// 			if _, err := url.Parse(envPsqlHost); err == nil { 
	// 				return &Options{
	// 					Addr:         envAddr,
	// 					BaseURL:      envBaseURL,
	// 					StorageFile:  storageFileValue,
	// 					DataBaseHost: envPsqlHost,
	// 				}, nil
	// 			}
	// 		}
	// 	}
	// }
}