package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paulochiaradia/smart-gondola-backend/internal/di"
	router "github.com/paulochiaradia/smart-gondola-backend/internal/interface/http"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/logger"
)

func main() {
	// 1. Setup (Config & Logger)
	cfg := config.Get()
	log := logger.New(cfg)
	log.Info("Inicializando Smart Gondola Backend...", "env", cfg.AppEnv)

	// 2. Dependências (Database & Services)
	container, cleanup, err := di.NewContainer(cfg)
	if err != nil {
		log.Error("Falha crítica ao inicializar dependências", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	// 3. Interface (HTTP Router)
	httpHandler := router.NewRouter(container)

	// 4. Server Start
	serverPort := fmt.Sprintf(":%s", cfg.ServerPort)
	server := &http.Server{
		Addr:    serverPort,
		Handler: httpHandler,
	}

	// --- Graceful Shutdown ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("Servidor HTTP rodando", "port", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Erro fatal no servidor HTTP", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	log.Info("Sinal de parada recebido. Desligando...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Erro ao desligar servidor forçadamente", "error", err)
	}

	log.Info("Servidor finalizado com sucesso.")
}
