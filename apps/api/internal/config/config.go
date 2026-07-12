package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/xromen/movietracker/internal/platform/database"
)

type Config struct {
	HTTP         HTTPConfig
	Database     DatabaseConfig
	TMDB         TMDBConfig
	JWT          JWTConfig
	RefreshToken RefreshTokenConfig
	Redis        RedisConfig
}

type HTTPConfig struct {
	Port         int
	ReadTimeout  int
	WriteTimeout int
}

type DatabaseConfig struct {
	Host             string
	Port             int
	User             string
	Password         string
	DBName           string
	SSLMode          string
	StatementTimeout time.Duration
}

type TMDBConfig struct {
	BaseURL       string
	ImagesBaseURL string
	BearerToken   string
	Timeout       time.Duration
}

type JWTConfig struct {
	Secret         string
	AccessTokenTTL time.Duration
}

type RefreshTokenConfig struct {
	RefreshTokenTTL time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Disabled bool
}

func (c DatabaseConfig) DSN() string {
	databaseURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   c.DBName,
	}

	query := databaseURL.Query()
	query.Set("sslmode", c.SSLMode)
	databaseURL.RawQuery = query.Encode()

	return databaseURL.String()
}

func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	redisDb, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	dbStatementTimeout, err := time.ParseDuration(getEnv("DB_STATEMENT_TIMEOUT", "15s"))
	if err != nil || dbStatementTimeout <= 0 {
		return nil, fmt.Errorf("invalid DB_STATEMENT_TIMEOUT")
	}

	tmdbTimeout, err := time.ParseDuration(getEnv("TMDB_TIMEOUT", "10s"))
	if err != nil || tmdbTimeout <= 0 {
		return nil, fmt.Errorf("invalid TMDB_TIMEOUT")
	}

	return &Config{
		HTTP: HTTPConfig{
			Port:         port,
			ReadTimeout:  10,
			WriteTimeout: 10,
		},
		Database: DatabaseConfig{
			Host:             getEnv("DB_HOST", "localhost"),
			Port:             dbPort,
			User:             getEnv("DB_USER", "postgres"),
			Password:         getEnv("DB_PASSWORD", ""),
			DBName:           getEnv("DB_NAME", "movietracker"),
			SSLMode:          getEnv("DB_SSLMODE", "disable"),
			StatementTimeout: dbStatementTimeout,
		},
		TMDB: TMDBConfig{
			BaseURL:       getEnv("TMDB_BASE_URL", "https://api.themoviedb.org/3"),
			ImagesBaseURL: getEnv("TMDB_IMAGES_BASE_URL", "https://api.themoviedb.org"),
			BearerToken:   getEnv("TMDB_BEARER_TOKEN", ""),
			Timeout:       tmdbTimeout,
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", "change-me-in-production"),
			AccessTokenTTL: 15 * time.Second,
		},
		RefreshToken: RefreshTokenConfig{
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDb,
			Disabled: getEnv("REDIS_DISABLED", "false") == "true",
		},
	}, nil
}

func (c DatabaseConfig) PoolConfig() database.Config {
	return database.Config{
		DSN:              c.DSN(),
		MaxConns:         25,
		MinConns:         5,
		MaxConnLifetime:  5 * time.Minute,
		MaxConnIdleTime:  10 * time.Minute,
		StatementTimeout: c.StatementTimeout,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
