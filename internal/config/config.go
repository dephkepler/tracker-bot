package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Telegram      Telegram
	PostreSQL     PgConfig
	AdminUsername string `env:"ADMIN_USERNAME" env-default:"alaamov"`
}
type PgConfig struct {
	Host    string `env:"HOST_DB"`
	Port    uint16 `env:"PORT_DB"`
	Name    string `env:"NAME_DB"`
	User    string `env:"USER_DB"`
	Pass    string `env:"PASSWORD_DB"`
	SSLMode string `env:"SSL_MODE" env-default:"disable"`
}
type Telegram struct {
	TelegramToken    string `env:"TELEGRAM_TOKEN"`
	TelegramBotDebug bool   `env:"TELEGRAM_BOT_DEBUG"`
}

const (
	envConfigPath     = "CONFIG_PATH"
	defaultConfigPath = ".env"
)

func ParseConfig() (*Config, error) {
	var cfg Config

	if os.Getenv("HOST_DB") != "" {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("read env: %w", err)
		}
		return &cfg, nil
	}

	// In local runs, load config from file path (default .env).
	path := os.Getenv(envConfigPath)
	if path == "" {
		path = defaultConfigPath
	}

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return &cfg, nil
}

func (c Config) PostgresDSN() string {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.PostreSQL.Host, c.PostreSQL.Port, c.PostreSQL.User, c.PostreSQL.Pass, c.PostreSQL.Name, c.PostreSQL.SSLMode,
	)
	return dsn
}

func (c Config) PostgresURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.PostreSQL.User,
		c.PostreSQL.Pass,
		c.PostreSQL.Host,
		c.PostreSQL.Port,
		c.PostreSQL.Name,
		c.PostreSQL.SSLMode,
	)
}
