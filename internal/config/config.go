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
	Addr    string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

func newOpts() (*Options, error) {

	envAddr := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	if envAddr != "" && envBaseURL != "" {
		if _, err := url.Parse("http://" + envAddr); err == nil {
			if _, err := url.Parse(envBaseURL); err == nil {
				return &Options{
					Addr:    envAddr,
					BaseURL: envBaseURL,
				}, nil
			}
		}
	}

	var addr = flag.String("a", "localhost:8080", "server host")
	var baseURL = flag.String("b", "localhost:8080", "value before short URL")
	flag.Parse()

	if _, err := url.Parse("https://" + *addr); err != nil {
		return nil, fmt.Errorf("incorrect parametr `-a` %s", *addr)
	}

	if _, err := url.Parse("https://" + *baseURL); err != nil {
		return nil, fmt.Errorf("incorrect parametr `-b` %s", *baseURL)
	}

	if !strings.HasPrefix(*baseURL, "http://") && !strings.HasPrefix(*baseURL, "https://") {
		*baseURL = "http://" + *baseURL
	}
	*baseURL = strings.TrimSuffix(*baseURL, "/")
	return &Options{
		Addr:    *addr,
		BaseURL: *baseURL,
	}, nil
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
