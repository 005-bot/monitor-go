package config

import (
	"fmt"
	"os"

	"github.com/go-core-fx/config"
)

const defaultPrefix = "bot-005"

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
	URL    string `koanf:"url"`
	Prefix string `koanf:"prefix"`
}

type scraperConfig struct {
	URL      string `koanf:"url"`
	Interval int    `koanf:"interval"`
}

type storageConfig struct {
	TTLDays int    `koanf:"ttl_days"`
	Prefix  string `koanf:"prefix"`
}

type publisherConfig struct {
	Prefix string `koanf:"prefix"`
}

type Config struct {
	HTTP      http            `koanf:"http"`
	Redis     redisConfig     `koanf:"redis"`
	Scraper   scraperConfig   `koanf:"scraper"`
	Storage   storageConfig   `koanf:"storage"`
	Publisher publisherConfig `koanf:"publisher"`
}

func Default() Config {
	//nolint:mnd // default values
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
			URL:    "redis://localhost:6379",
			Prefix: defaultPrefix,
		},

		Scraper: scraperConfig{
			URL:      "http://93.92.65.26/aspx/Gorod.htm",
			Interval: 300,
		},

		Storage: storageConfig{
			TTLDays: 5,
			Prefix:  defaultPrefix,
		},

		Publisher: publisherConfig{
			Prefix: defaultPrefix,
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
