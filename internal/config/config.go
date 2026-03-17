package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Opts  *Options
	Token *Token
}

type Options struct {
	Addr         string `json:"server_address" env:"SERVER_ADDRESS"`
	BaseURL      string `json:"base_url" env:"BASE_URL"`
	StorageFile  string `json:"file_storage_path" env:"FILE_STORAGE_PATH"`
	DataBaseHost string `json:"database_dsn" env:"DATABASE_DSN"`

	AuditFile string `env:"AUDIT_FILE"`
	AuditURL  string `env:"AUDIT_URL"`
	HTTPS     bool   `json:"enable_https" env:"ENABLE_HTTPS"`
	Subnet    string `json:"trusted_subnet" env:"TRUSTED_SUBNET"`
}

type Token struct {
	SecretKey  string        `env:"SECRET_KEY" default:"super-secret-key"`
	CookieName string        `env:"COOKIE_NAME" default:"url_shortener_session"`
	CookieTTL  time.Duration `env:"COOKIE_TTL" default:"24h"`
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
	var storageFile = flag.String("f", "", "file for save data")
	var psqlHost = flag.String("d", "", "psql data")
	var auditFile = flag.String("audit-file", "", "audit file save events")
	var auditURL = flag.String("audit-url", "", "audit url send events")
	var https = flag.Bool("s", false, "run https")
	var subnet = flag.String("t", "", "subnet")
	var configJSON string
	flag.StringVar(&configJSON, "c", "", "config file (short)")
	flag.StringVar(&configJSON, "config", "", "config file (long)")

	flag.Parse()

	if opts.Addr == "" {
		opts.Addr = *addr
	}
	if opts.BaseURL == "" {
		opts.BaseURL = *baseURL
	}
	if opts.StorageFile == "" {
		opts.StorageFile = *storageFile
	}
	if opts.DataBaseHost == "" {
		opts.DataBaseHost = *psqlHost
	}
	if opts.Subnet == "" {
		opts.Subnet = *subnet
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
	opts.HTTPS = *https

	parseAuditFields(opts, auditFile, auditURL)
	if cfgFile := getConfigFilePath(&configJSON); cfgFile != nil {
		if err := setConfigFromFile(*cfgFile, opts); err != nil {
			logrus.Warn(err)
		}
	}
	return opts, nil
}

func parseEnv() (*Options, bool) {
	envAddr := os.Getenv("SERVER_ADDRESS")
	envBaseURL := os.Getenv("BASE_URL")
	envStorageFile := os.Getenv("FILE_STORAGE_PATH")
	envPsqlDsn := os.Getenv("DATABASE_DSN")

	envAiditFile := os.Getenv("AUDIT_FILE")
	envAiditURL := os.Getenv("AUDIT_URL")
	envHTTPS := os.Getenv("ENABLE_HTTPS")
	envSubnet := os.Getenv("TRUSTED_SUBNET")

	opts := &Options{}
	if envStorageFile != "" {
		opts.StorageFile = envStorageFile
	}
	if envAiditFile != "" {
		opts.AuditFile = envAiditFile
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
	if _, err := url.Parse(envAiditURL); err == nil {
		opts.AuditURL = envAiditURL
	}
	if envHTTPS != "" {
		opts.HTTPS = true
	}
	if envSubnet != "" {
		opts.Subnet = envSubnet
	}

	sucessAll := opts.Addr != "" && opts.BaseURL != "" && opts.DataBaseHost != ""
	return opts, sucessAll
}

func parseAuditFields(opts *Options, auditFilePath, auditURL *string) {
	if opts.AuditFile == "" && auditFilePath != nil {
		opts.AuditFile = *auditFilePath
	}
	if opts.AuditURL == "" && auditURL != nil {
		if _, err := url.Parse(*auditURL); err == nil {
			opts.AuditURL = *auditURL
		}
	}
}

// getConfigFilePath returns the path to the config file specified by the -c flag or the CONFIG environment variable.
func getConfigFilePath(configJSON *string) *string {
	cfgFile := os.Getenv("CONFIG")
	if cfgFile != "" {
		return &cfgFile
	}

	return configJSON
}

func setConfigFromFile(path string, opts *Options) error {
	var optFromFile Options

	data, err := os.ReadFile(path)
	if err != nil {
		logrus.Error(err)
		return err
	}

	err = json.Unmarshal(data, &optFromFile)
	if err != nil {
		logrus.Error(err)
		return err
	}

	if flag.Lookup("a") == nil {
		opts.Addr = optFromFile.Addr
	}

	if flag.Lookup("b") == nil {
		opts.BaseURL = optFromFile.BaseURL
	}

	if flag.Lookup("f") == nil {
		opts.StorageFile = optFromFile.StorageFile
	}

	if flag.Lookup("d") == nil {
		opts.DataBaseHost = optFromFile.DataBaseHost
	}

	if flag.Lookup("s") == nil {
		opts.HTTPS = optFromFile.HTTPS
	}

	return nil
}
