package config

import (
	"flag"
	"fmt"
	"net/url"
)

type Config struct {
	Opts *Options
}

type Options struct {
	Addr    string
	BaseURL string
}

func newOpts() (*Options, error) {
	var addr = flag.String("a", "http://localhost:8080", "server host")
	var baseURL = flag.String("b", "http://localhost:8080", "value before short URL")
	flag.Parse()

	parsedURL, err := url.Parse(*addr)
	if err != nil {
		return nil, fmt.Errorf("incorrect parametr `-a` %s", *addr)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("incorrect parametr `-a` %s", *addr)
	}

	parsedURL, err = url.Parse(*baseURL)
	if err != nil {
		return nil, fmt.Errorf("incorrect parametr `-b` %s", *baseURL)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("incorrect parametr `-a` %s", *baseURL)
	}
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
