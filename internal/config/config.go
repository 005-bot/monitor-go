package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/go-core-fx/config"
)

var (
	ErrInvalidConfig  = errors.New("invalid configuration")
	errMsgURLRequired = errors.New("must be a valid URL")
	errMsgPositive    = errors.New("must be greater than 0")
	errMsgNotEmpty    = errors.New("must not be empty")
)

const (
	defaultScraperIntervalSec = 300 // 5 minutes in seconds
	defaultScraperTimeout     = time.Minute
	defaultStorageTTLDays     = 5
)

type http struct {
	Address     string   `koanf:"address"`
	ProxyHeader string   `koanf:"proxy_header"`
	Proxies     []string `koanf:"proxies"`

	OpenAPI openAPIConfig `koanf:"openapi"`
}

type openAPIConfig struct {
	Enabled    bool   `koanf:"enabled"`
	PublicHost string `koanf:"public_host"`
	PublicPath string `koanf:"public_path"`
}

type redisConfig struct {
	URL string `koanf:"url"`
}

type scraperConfig struct {
	URL      string        `koanf:"url"`
	Interval int           `koanf:"interval"`
	Timeout  time.Duration `koanf:"timeout"`
}

type storageConfig struct {
	TTLDays int    `koanf:"ttl_days"`
	Prefix  string `koanf:"prefix"`
}

type publisherConfig struct {
	Prefix string `koanf:"prefix"`
}

type parserConfig struct {
	AddressDBPath string `koanf:"address_db_path"`
}

type Config struct {
	HTTP      http            `koanf:"http"`
	Redis     redisConfig     `koanf:"redis"`
	Scraper   scraperConfig   `koanf:"scraper"`
	Storage   storageConfig   `koanf:"storage"`
	Publisher publisherConfig `koanf:"publisher"`
	Parser    parserConfig    `koanf:"parser"`
}

func Default() Config {
	return Config{
		HTTP: http{
			Address:     "127.0.0.1:3000",
			ProxyHeader: "X-Forwarded-For",
			Proxies:     []string{},
			OpenAPI: openAPIConfig{
				Enabled:    true,
				PublicHost: "",
				PublicPath: "",
			},
		},
		Redis: redisConfig{
			URL: "redis://localhost:6379",
		},
		Scraper: scraperConfig{
			URL:      "http://93.92.65.26/aspx/Gorod.htm",
			Interval: defaultScraperIntervalSec,
			Timeout:  defaultScraperTimeout,
		},
		Storage: storageConfig{
			TTLDays: defaultStorageTTLDays,
			Prefix:  "bot-005",
		},
		Publisher: publisherConfig{
			Prefix: "bot-005",
		},
		Parser: parserConfig{
			AddressDBPath: "",
		},
	}
}

func New() (Config, error) {
	cfg := Default()

	options := []config.Option{}
	if yamlPath := os.Getenv("CONFIG_PATH"); yamlPath != "" {
		options = append(options, config.WithLocalYAML(yamlPath))
	}

	if err := config.Load(&cfg, options...); err != nil {
		return Config{}, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error

	if _, err := url.Parse(c.Redis.URL); c.Redis.URL == "" || err != nil {
		errs = append(errs, fmt.Errorf("redis.url: %w", errMsgURLRequired))
	}

	if _, err := url.Parse(c.Scraper.URL); c.Scraper.URL == "" || err != nil {
		errs = append(errs, fmt.Errorf("scraper.url: %w", errMsgURLRequired))
	}

	if c.Scraper.Interval <= 0 {
		errs = append(errs, fmt.Errorf("scraper.interval: %w", errMsgPositive))
	}

	if c.Storage.TTLDays <= 0 {
		errs = append(errs, fmt.Errorf("storage.ttl_days: %w", errMsgPositive))
	}

	if c.Storage.Prefix == "" {
		errs = append(errs, fmt.Errorf("storage.prefix: %w", errMsgNotEmpty))
	}

	if c.Publisher.Prefix == "" {
		errs = append(errs, fmt.Errorf("publisher.prefix: %w", errMsgNotEmpty))
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w", errors.Join(append([]error{ErrInvalidConfig}, errs...)...))
	}

	return nil
}
