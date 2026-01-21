package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
)

// New configura o logger global. Deve ser chamado no início do main.go
func New(cfg *config.Config) *slog.Logger {
	var handler slog.Handler

	// Define o nível de log (Debug em dev, Info em prod)
	opts := &slog.HandlerOptions{
		Level: logLevel(cfg.AppEnv),
	}

	// Define o formato (JSON para máquinas, Texto para humanos)
	if strings.ToLower(cfg.LogFormat) == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)

	// Define como logger padrão global para que bibliotecas de terceiros também usem
	slog.SetDefault(logger)

	return logger
}

func logLevel(env string) slog.Level {
	if strings.ToLower(env) == "production" {
		return slog.LevelInfo
	}
	return slog.LevelDebug
}
