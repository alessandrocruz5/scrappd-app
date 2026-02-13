package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	MLService  MLServiceConfig
	Storage    StorageConfig
	CloudTasks CloudTasksConfig
	JWT        JWTConfig
	Email      EmailConfig
	App        AppConfig
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	Environment     string        `mapstructure:"environment"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	DSN      string
}

type RedisConfig struct {
	URL      string
	Host     string
	Port     string
	Password string
	DB       int
}

type MLServiceConfig struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
}

type CloudTasksConfig struct {
	Enabled             bool
	ProjectID           string
	Location            string
	QueueID             string
	ServiceURL          string
	ServiceAccountEmail string
}

type JWTConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	VerifyTokenSecret  string
	VerifyTokenExpiry  time.Duration
}

type EmailConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	FromName        string
	FromEmail       string
	UseTLS          bool
	UseStartTLS     bool
	RequireStartTLS bool
	SkipTLSVerify   bool
}

type AppConfig struct {
	BaseURL            string
	BypassUsageLimits  bool
	InternalTaskSecret string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("PORT", getEnv("SERVER_PORT", "8080")),
			Environment:     getEnv("ENVIRONMENT", "development"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "scrappd_app"),
			Password: getEnv("DB_PASSWORD", "usTiCr$9S%B5u2"),
			DBName:   getEnv("DB_NAME", "scrappd"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", ""),
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		MLService: MLServiceConfig{
			BaseURL:    getEnv("ML_SERVICE_URL", "http://localhost:8000"),
			Timeout:    getDurationEnv("ML_SERVICE_TIMEOUT", 120*time.Second),
			MaxRetries: getIntEnv("ML_SERVICE_MAX_RETRIES", 3),
			RetryDelay: getDurationEnv("ML_SERVICE_RETRY_DELAY", 2*time.Second),
		},
		Storage: StorageConfig{
			Endpoint:        getEnv("STORAGE_ENDPOINT", ""),
			AccessKeyID:     getEnv("STORAGE_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("STORAGE_SECRET_ACCESS_KEY", ""),
			BucketName:      getEnv("STORAGE_BUCKET_NAME", "scrappd-images"),
			Region:          getEnv("STORAGE_REGION", "auto"),
		},
		CloudTasks: CloudTasksConfig{
			Enabled:             getBoolEnv("CLOUD_TASKS_ENABLED", false),
			ProjectID:           getEnv("CLOUD_TASKS_PROJECT_ID", ""),
			Location:            getEnv("CLOUD_TASKS_LOCATION", ""),
			QueueID:             getEnv("CLOUD_TASKS_QUEUE_ID", ""),
			ServiceURL:          getEnv("CLOUD_TASKS_SERVICE_URL", ""),
			ServiceAccountEmail: getEnv("CLOUD_TASKS_SERVICE_ACCOUNT_EMAIL", ""),
		},
		JWT: JWTConfig{
			AccessTokenSecret:  getEnv("JWT_ACCESS_SECRET", "your-secret-key-change-in-production"),
			RefreshTokenSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-change-in-production"),
			AccessTokenExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			VerifyTokenSecret:  getEnv("JWT_VERIFY_SECRET", "your-verify-secret-change-in-production"),
			VerifyTokenExpiry:  getDurationEnv("JWT_VERIFY_EXPIRY", 24*time.Hour),
		},
		Email: EmailConfig{
			Host:            getEnv("SMTP_HOST", ""),
			Port:            getIntEnv("SMTP_PORT", 587),
			Username:        getEnv("SMTP_USERNAME", ""),
			Password:        getEnv("SMTP_PASSWORD", ""),
			FromName:        getEnv("SMTP_FROM_NAME", "Scrapp'd"),
			FromEmail:       getEnv("SMTP_FROM_EMAIL", ""),
			UseTLS:          getBoolEnv("SMTP_USE_TLS", false),
			UseStartTLS:     getBoolEnv("SMTP_USE_STARTTLS", true),
			RequireStartTLS: getBoolEnv("SMTP_REQUIRE_STARTTLS", false),
			SkipTLSVerify:   getBoolEnv("SMTP_SKIP_TLS_VERIFY", false),
		},
		App: AppConfig{
			BaseURL:            getEnv("APP_BASE_URL", "http://localhost:3000"),
			BypassUsageLimits:  getBoolEnv("BYPASS_USAGE_LIMITS", false),
			InternalTaskSecret: getEnv("INTERNAL_TASK_SECRET", ""),
		},
	}

	// Build database DSN
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			config.Database.Host,
			config.Database.Port,
			config.Database.User,
			config.Database.Password,
			config.Database.DBName,
			config.Database.SSLMode,
		)
	}
	config.Database.DSN = dsn

	if config.Server.Environment != "development" {
		if config.JWT.AccessTokenSecret == "your-secret-key-change-in-production" ||
			config.JWT.RefreshTokenSecret == "your-refresh-secret-change-in-production" {
			return nil, fmt.Errorf("JWT secrets must be set for non-development environments")
		}
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
