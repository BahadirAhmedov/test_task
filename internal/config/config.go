// Package config загружает конфигурацию приложения из переменных окружения.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
)

type Config struct {
	App      App
	Telegram Telegram
	LLM      LLM
	Postgres Postgres
}

type App struct {
	HTTPAddr string
}

type Telegram struct {
	BotToken string
}

type LLM struct {
	BaseURL string
	APIKey  string
	Model   string
}

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
	SSLMode  string
}

// DSN возвращает строку подключения вида postgres://user:pass@host:port/db?sslmode=...
func (p Postgres) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     net.JoinHostPort(p.Host, p.Port),
		Path:     p.DB,
		RawQuery: url.Values{"sslmode": {p.SSLMode}}.Encode(),
	}
	return u.String()
}

// Load читает конфиг из окружения и проверяет, что обязательные поля заданы.
func Load() (*Config, error) {
	var l loader

	cfg := &Config{
		App: App{
			HTTPAddr: l.optional("HTTP_ADDR", ":8080"),
		},
		Telegram: Telegram{
			BotToken: l.required("TELEGRAM_BOT_TOKEN"),
		},
		LLM: LLM{
			BaseURL: l.required("LLM_BASE_URL"),
			APIKey:  l.required("LLM_API_KEY"),
			Model:   l.required("LLM_MODEL"),
		},
		Postgres: Postgres{
			Host:     l.optional("POSTGRES_HOST", "localhost"),
			Port:     l.optional("POSTGRES_PORT", "5432"),
			User:     l.required("POSTGRES_USER"),
			Password: l.required("POSTGRES_PASSWORD"),
			DB:       l.required("POSTGRES_DB"),
			SSLMode:  l.optional("POSTGRES_SSLMODE", "disable"),
		},
	}

	if err := errors.Join(l.errs...); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}

// loader копит ошибки, чтобы сообщить обо всех незаданных переменных сразу.
type loader struct {
	errs []error
}

func (l *loader) required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		l.errs = append(l.errs, fmt.Errorf("%s is required", key))
	}
	return v
}

func (l *loader) optional(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
