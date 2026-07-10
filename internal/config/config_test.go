package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/005-bot/monitor-go/internal/config"
)

func TestDefaultIsValid(t *testing.T) {
	cfg := config.Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Default() Validate() = %v, want nil", err)
	}
}

func TestValidate_Errors(t *testing.T) {
	cfg := config.Default()

	cfg.Redis.URL = ""
	cfg.Scraper.URL = ""
	cfg.Scraper.Interval = 0
	cfg.Storage.TTLDays = 0
	cfg.Storage.Prefix = ""
	cfg.Publisher.Prefix = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error")
	}

	var target interface{ Unwrap() []error }
	if !errors.As(err, &target) {
		t.Fatal("Validate() error should be a joined error")
	}

	errs := target.Unwrap()
	if len(errs) != 7 {
		t.Fatalf("expected 7 validation errors, got %d: %v", len(errs), errs)
	}
	if !errors.Is(errs[0], config.ErrInvalidConfig) {
		t.Error("first error should be ErrInvalidConfig")
	}

	msg := err.Error()
	for _, snippet := range []string{
		"redis.url",
		"scraper.url",
		"scraper.interval",
		"storage.ttl_days",
		"storage.prefix",
		"publisher.prefix",
	} {
		if !strings.Contains(msg, snippet) {
			t.Errorf("error should mention %q", snippet)
		}
	}

	if !errors.Is(err, config.ErrInvalidConfig) {
		t.Error("Validate() error should wrap ErrInvalidConfig")
	}
}

func TestValidate_InvalidRedisURL(t *testing.T) {
	cfg := config.Default()
	cfg.Redis.URL = "://invalid"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error")
	}
	if !strings.Contains(err.Error(), "redis.url") {
		t.Errorf("error should mention redis.url")
	}
}

func TestValidate_InvalidScraperURL(t *testing.T) {
	cfg := config.Default()
	cfg.Scraper.URL = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want error")
	}
	if !strings.Contains(err.Error(), "scraper.url") {
		t.Errorf("error should mention scraper.url")
	}
}
