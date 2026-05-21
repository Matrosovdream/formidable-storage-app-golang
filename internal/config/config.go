package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
	Cache CacheConfig
	Queue QueueConfig
	Auth  AuthConfig
	Mail  MailConfig
	Log   LogConfig
	CORS  CORSConfig
}

type AppConfig struct {
	Name     string
	Env      string
	Port     int
	URL      string
	Debug    bool
	Timezone string
}

type CORSConfig struct {
	AllowedOrigins string
	AllowedMethods string
	AllowedHeaders string
}

type DBConfig struct {
	Host         string
	Port         int
	Database     string
	Username     string
	Password     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type CacheConfig struct {
	Driver string
	TTL    int
	Prefix string
}

type QueueConfig struct {
	Connection string
	Stream     string
	StatsKey   string
}

type AuthConfig struct {
	SessionLifetimeMinutes int
	TokenLifetimeMinutes   int
	BcryptCost             int
}

type MailConfig struct {
	Driver   string
	Host     string
	Port     int
	From     string
	FromName string
}

type LogConfig struct {
	Level string
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("json")
	v.AddConfigPath(".")
	v.AddConfigPath("/src")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := &Config{
		App: AppConfig{
			Name:     v.GetString("APP_NAME"),
			Env:      v.GetString("APP_ENV"),
			Port:     v.GetInt("APP_PORT"),
			URL:      v.GetString("APP_URL"),
			Debug:    v.GetBool("APP_DEBUG"),
			Timezone: v.GetString("APP_TIMEZONE"),
		},
		DB: DBConfig{
			Host:         v.GetString("DB_HOST"),
			Port:         v.GetInt("DB_PORT"),
			Database:     v.GetString("DB_DATABASE"),
			Username:     v.GetString("DB_USERNAME"),
			Password:     v.GetString("DB_PASSWORD"),
			SSLMode:      v.GetString("DB_SSLMODE"),
			MaxOpenConns: v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns: v.GetInt("DB_MAX_IDLE_CONNS"),
		},
		Redis: RedisConfig{
			Host:     v.GetString("REDIS_HOST"),
			Port:     v.GetInt("REDIS_PORT"),
			Password: v.GetString("REDIS_PASSWORD"),
			DB:       v.GetInt("REDIS_DB"),
		},
		Cache: CacheConfig{
			Driver: v.GetString("CACHE_SERVICE_DRIVER"),
			TTL:    v.GetInt("CACHE_SERVICE_TTL"),
			Prefix: v.GetString("CACHE_SERVICE_PREFIX"),
		},
		Queue: QueueConfig{
			Connection: v.GetString("QUEUE_CONNECTION"),
			Stream:     v.GetString("QUEUE_STREAM"),
			StatsKey:   v.GetString("QUEUE_STATS_KEY"),
		},
		Auth: AuthConfig{
			SessionLifetimeMinutes: v.GetInt("SESSION_LIFETIME"),
			TokenLifetimeMinutes:   v.GetInt("TOKEN_LIFETIME"),
			BcryptCost:             v.GetInt("BCRYPT_ROUNDS"),
		},
		Mail: MailConfig{
			Driver:   v.GetString("MAIL_MAILER"),
			Host:     v.GetString("MAIL_HOST"),
			Port:     v.GetInt("MAIL_PORT"),
			From:     v.GetString("MAIL_FROM_ADDRESS"),
			FromName: v.GetString("MAIL_FROM_NAME"),
		},
		Log: LogConfig{
			Level: v.GetString("LOG_LEVEL"),
		},
		CORS: CORSConfig{
			AllowedOrigins: v.GetString("CORS_ALLOWED_ORIGINS"),
			AllowedMethods: v.GetString("CORS_ALLOWED_METHODS"),
			AllowedHeaders: v.GetString("CORS_ALLOWED_HEADERS"),
		},
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("APP_NAME", "FormidableStorageApp")
	v.SetDefault("APP_ENV", "local")
	v.SetDefault("APP_DEBUG", true)
	v.SetDefault("APP_PORT", 8080)
	v.SetDefault("APP_TIMEZONE", "UTC")

	v.SetDefault("DB_HOST", "postgres")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_DATABASE", "app")
	v.SetDefault("DB_USERNAME", "laravel")
	v.SetDefault("DB_PASSWORD", "secret")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 5)

	v.SetDefault("REDIS_HOST", "redis")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_DB", 0)

	v.SetDefault("CACHE_SERVICE_DRIVER", "redis")
	v.SetDefault("CACHE_SERVICE_TTL", 3600)
	v.SetDefault("CACHE_SERVICE_PREFIX", "fsa:")

	v.SetDefault("QUEUE_CONNECTION", "redis")
	v.SetDefault("QUEUE_STREAM", "queues:default")
	v.SetDefault("QUEUE_STATS_KEY", "queue-stats")

	v.SetDefault("SESSION_LIFETIME", 120)
	v.SetDefault("TOKEN_LIFETIME", 0)
	v.SetDefault("BCRYPT_ROUNDS", 12)

	v.SetDefault("LOG_LEVEL", "info")

	v.SetDefault("CORS_ALLOWED_ORIGINS", "*")
	v.SetDefault("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	v.SetDefault("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization,X-Requested-With")
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}
