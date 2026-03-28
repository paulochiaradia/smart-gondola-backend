package config

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

var (
	cfgInstance *Config
	once        sync.Once
)

type Config struct {
	AppEnv             string
	ServerPort         string
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ServerIdleTimeout  time.Duration
	LogFormat          string
	BaseURL            string // Importante para montar URLs de imagens (Avatar)

	// --- Database ---
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	// --- Security ---
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshExpiration time.Duration

	// --- Storage (Para Avatars e Imagens de Produtos) ---
	StorageDriver string // 'local', 's3'
	AWSBucket     string
	AWSRegion     string
	AWSAccessKey  string
	AWSSecretKey  string

	// --- Mobile Notifications (Firebase/FCM) ---
	FirebaseCredsFile string // Caminho para o .json do Google
}

func Get() *Config {
	once.Do(func() {
		_ = godotenv.Load()

		cfgInstance = &Config{
			AppEnv:             getEnv("APP_ENV", "development"),
			ServerPort:         getEnv("SERVER_PORT", "8080"),
			ServerReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			ServerWriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 20*time.Second),
			ServerIdleTimeout:  getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			LogFormat:          getEnv("LOG_FORMAT", "json"),
			BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),

			DBHost: getEnv("DB_HOST", "127.0.0.1"),
			DBPort: getEnv("DB_PORT", "5432"),
			DBUser: getEnv("DB_USER", "postgres"),
			DBPass: getEnv("DB_PASS", "postgres"),
			DBName: getEnv("DB_NAME", "smartgondola"),

			JWTSecret:         getEnv("JWT_SECRET", "default_secret"),
			JWTExpiration:     time.Hour * 24,      // 1 dia (Access Token)
			RefreshExpiration: time.Hour * 24 * 30, // 30 dias (Mobile não desloga fácil)

			// Storage Defaults (Local para dev)
			StorageDriver: getEnv("STORAGE_DRIVER", "local"),
			AWSBucket:     getEnv("AWS_BUCKET", ""),
			AWSRegion:     getEnv("AWS_REGION", "us-east-1"),
			AWSAccessKey:  getEnv("AWS_ACCESS_KEY_ID", ""),
			AWSSecretKey:  getEnv("AWS_SECRET_ACCESS_KEY", ""),

			// Notifications
			FirebaseCredsFile: getEnv("FIREBASE_CREDENTIALS", "firebase-service-account.json"),
		}
	})
	return cfgInstance
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}

	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return fallback
}
