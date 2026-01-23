package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/application/usecase"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/infrastructure/repository"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/interface/http/handler"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/database"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/logger"
)

func main() {
	// 1. Carrega Configuração
	cfg := config.Get()

	// 2. Inicializa Logger
	log := logger.New(cfg)
	log.Info("Inicializando Smart Gondola Backend...", "env", cfg.AppEnv)

	// 3. Conecta no Banco de Dados
	db, err := database.NewPostgres(cfg)
	if err != nil {
		log.Error("Falha crítica ao conectar no banco", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("Conexão com Postgres estabelecida")

	// 4. Injeção de Dependência (Wiring)
	// User Module
	userRepo := repository.NewUserRepository(db)
	userUseCase := usecase.NewUserUseCase(userRepo)
	userHandler := handler.NewUserHandler(userUseCase)

	// 5. Configura Roteador (Chi)
	r := chi.NewRouter()

	// Middlewares Globais
	r.Use(middleware.RequestID) // Gera ID único para cada request (bom para rastreio)
	r.Use(middleware.RealIP)    // Pega o IP real do usuário
	r.Use(middleware.Logger)    // Loga cada requisição
	r.Use(middleware.Recoverer) // Evita que o server caia se der Panic
	r.Use(middleware.Timeout(60 * time.Second))

	// Rota de Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API V1
	r.Route("/api/v1", func(r chi.Router) {
		userHandler.RegisterRoutes(r)
	})

	// 6. Inicia o Servidor (Com Graceful Shutdown)
	serverPort := fmt.Sprintf(":%s", cfg.ServerPort)
	server := &http.Server{
		Addr:    serverPort,
		Handler: r,
	}

	// Canal para escutar sinais do SO (CTRL+C, Docker Stop)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("Servidor HTTP rodando", "port", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Erro fatal no servidor HTTP", "error", err)
			os.Exit(1)
		}
	}()

	// Bloqueia execução esperando sinal de parada
	<-stop
	log.Info("Sinal de parada recebido. Desligando...")

	// Contexto com timeout para terminar requests em andamento
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Erro ao desligar servidor forçadamente", "error", err)
	}

	log.Info("Servidor finalizado com sucesso.")
}
