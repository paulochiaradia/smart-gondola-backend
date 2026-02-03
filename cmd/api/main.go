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
	// Alias "router" para evitar conflito com pacote "net/http"
	router "github.com/paulochiaradia/smart-gondola-backend/internal/interface/http"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/database"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/logger"
)

func main() {
	// 1. Setup (Config & Logger)
	cfg := config.Get()
	log := logger.New(cfg)
	log.Info("Inicializando Smart Gondola Backend...", "env", cfg.AppEnv)

	// Monta a string de conexão para o Migrator
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	log.Info("Verificando migrações de banco de dados...")
	if err := database.RunMigrations(dbURL); err != nil {
		log.Error("Falha crítica ao rodar migrations", "error", err)
		// Se o banco não estiver atualizado, a aplicação não deve subir.
		os.Exit(1)
	}
	// ---------------------------

	// 2. Dependências (Database Connection Pool & Container DI)
	container, cleanup, err := di.NewContainer(cfg)
	if err != nil {
		log.Error("Falha crítica ao inicializar dependências", "error", err)
		os.Exit(1)
	}
	// Garante que o banco fecha quando o main morrer
	defer cleanup()

	// 3. Interface (HTTP Router)
	// A main delega a criação do http.Handler para a camada de interface
	httpHandler := router.NewRouter(container)

	// 4. Server Start
	serverPort := fmt.Sprintf(":%s", cfg.ServerPort)
	server := &http.Server{
		Addr:    serverPort,
		Handler: httpHandler,
	}

	// --- Graceful Shutdown (Desligamento Gracioso) ---
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("Servidor HTTP rodando", "port", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Erro fatal no servidor HTTP", "error", err)
			os.Exit(1)
		}
	}()

	// Bloqueia a execução aqui esperando CTRL+C ou Stop do Docker
	<-stop
	log.Info("Sinal de parada recebido. Desligando...")

	// Dá 5 segundos para as requisições atuais terminarem antes de matar
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Erro ao desligar servidor forçadamente", "error", err)
	}

	log.Info("Servidor finalizado com sucesso.")
}
