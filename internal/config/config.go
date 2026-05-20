package config

import (
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config хранит всю конфигурацию приложения.
type Config struct {
	Port string `env:"HTTP_PORT" envDefault:"8080"`
	DB   DBConfig
	JWT  JWTConfig
}

type DBConfig struct {
	Host     string `env:"DB_HOST,required"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
	Name     string `env:"DB_NAME,required"`
	SSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`
	MaxConns int    `env:"DB_MAX_CONNS" envDefault:"10"`
}

type JWTConfig struct {
	Secret          string        `env:"JWT_SECRET,required"`
	AccessTokenTTL  time.Duration `env:"JWT_ACCESS_TTL"  envDefault:"15m"`
	RefreshTokenTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`
}

// Load читает .env (если есть) и парсит конфиг из переменных окружения.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение: .env файл не найден, используются системные переменные окружения")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфига: %w", err)
	}
	return cfg, nil
}

// ConnectionString возвращает строку подключения к PostgreSQL.
func (c DBConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode, c.MaxConns,
	)
}
