// Package config loads runtime configuration from environment variables.
//
// Local development reads a .env file (see .env.example); production is expected
// to inject real environment variables. Secrets never live in the repository.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env     string // "dev" | "staging" | "prod"
	HTTP    HTTPConfig
	DB      DBConfig
	Redis   RedisConfig
	Auth    AuthConfig
	Lava    LavaConfig
	Storage StorageConfig
	Billing BillingConfig
}

type HTTPConfig struct {
	Port           string
	AllowedOrigins []string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.Host, c.User, c.Password, c.Name, c.Port, c.SSLMode,
	)
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type AuthConfig struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	CookieDomain    string
	CookieSecure    bool
}

// LavaConfig holds credentials for lava.top (payment acceptance only).
type LavaConfig struct {
	APIKey     string
	WebhookKey string // shared secret to verify incoming webhooks
	BaseURL    string
	OfferID    string
}

// BillingConfig holds the platform's money parameters.
type BillingConfig struct {
	CommissionBps  int64 // platform cut of each released milestone, basis points
	PayoutMinCents int64
}

// StorageConfig points at an S3-compatible object store (MinIO locally).
type StorageConfig struct {
	Endpoint      string // host:port, no scheme
	AccessKey     string
	SecretKey     string
	Bucket        string
	UseSSL        bool
	PublicBaseURL string // where objects are served from, e.g. http://localhost:9000/ipw
}

// Load reads configuration. It loads .env if present (dev convenience) and then
// resolves every value from the environment, failing fast on missing required
// secrets.
func Load() (Config, error) {
	_ = godotenv.Load() // absent in prod; not an error

	cfg := Config{
		Env: env("APP_ENV", "dev"),
		HTTP: HTTPConfig{
			Port:           env("HTTP_PORT", "5000"),
			AllowedOrigins: splitAndTrim(env("HTTP_ALLOWED_ORIGINS", "http://localhost:3000")),
		},
		DB: DBConfig{
			Host:     env("DB_HOST", "localhost"),
			Port:     env("DB_PORT", "5432"),
			User:     env("DB_USER", "ipw"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     env("DB_NAME", "ipw"),
			SSLMode:  env("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:     env("REDIS_ADDR", "localhost:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       envInt("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			JWTSecret:       os.Getenv("JWT_SECRET"),
			AccessTokenTTL:  envDuration("AUTH_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: envDuration("AUTH_REFRESH_TTL", 720*time.Hour),
			CookieDomain:    env("AUTH_COOKIE_DOMAIN", "localhost"),
			CookieSecure:    envBool("AUTH_COOKIE_SECURE", false),
		},
		Lava: LavaConfig{
			APIKey:     os.Getenv("LAVA_API_KEY"),
			WebhookKey: os.Getenv("LAVA_WEBHOOK_KEY"),
			BaseURL:    env("LAVA_BASE_URL", "https://gate.lava.top"),
			OfferID:    os.Getenv("LAVA_OFFER_ID"),
		},
		Billing: BillingConfig{
			CommissionBps:  int64(envInt("PLATFORM_COMMISSION_BPS", 1000)), // 10%
			PayoutMinCents: int64(envInt("PAYOUT_MIN_CENTS", 100000)),      // 1000 RUB
		},
		Storage: StorageConfig{
			Endpoint:      env("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKey:     env("STORAGE_ACCESS_KEY", "ipw"),
			SecretKey:     env("STORAGE_SECRET_KEY", "ipw_local_dev"),
			Bucket:        env("STORAGE_BUCKET", "ipw"),
			UseSSL:        envBool("STORAGE_USE_SSL", false),
			PublicBaseURL: env("STORAGE_PUBLIC_BASE_URL", "http://localhost:9000/ipw"),
		},
	}

	var missing []string
	if cfg.DB.Password == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		missing = append(missing, "JWT_SECRET (min 32 chars)")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing/invalid required config: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func (c Config) IsProd() bool { return c.Env == "prod" }

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
