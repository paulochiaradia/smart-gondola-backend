package config

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// cfgInstance é a variável privada que segura a única instância (Singleton)
var (
	cfgInstance *Config
	once        sync.Once
)

// Config define a estrutura de todas as variáveis de ambiente
type Config struct {
	AppEnv     string // 'development' ou 'production'
	ServerPort string
	LogFormat  string // 'json' ou 'text'

	// Database
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	// Security
	JWTSecret string
}

// Get carrega a configuração na primeira chamada e retorna a instância
func Get() *Config {
	once.Do(func() {
		// Tenta carregar o .env. Em produção (Docker/K8s), isso pode falhar silenciosamente
		// pois as vars virão do ambiente real, o que é o comportamento esperado.
		_ = godotenv.Load()

		cfgInstance = &Config{
			AppEnv:     getEnv("APP_ENV", "development"),
			ServerPort: getEnv("SERVER_PORT", "8080"),
			LogFormat:  getEnv("LOG_FORMAT", "json"), // Cloud Native prefere JSON

			DBHost: getEnv("DB_HOST", "localhost"),
			DBPort: getEnv("DB_PORT", "5432"),
			DBUser: getEnv("DB_USER", "postgres"),
			DBPass: getEnv("DB_PASS", "postgres"),
			DBName: getEnv("DB_NAME", "shelflink"),

			JWTSecret: getEnv("JWT_SECRET", "default_inseguro_mude_em_prod"),
		}
	})
	return cfgInstance
}

// Helper para ler env com valor padrão
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
