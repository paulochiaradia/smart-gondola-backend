package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/paulochiaradia/smart-gondola-backend/internal/di"
)

// NewRouter configura e retorna o handler HTTP principal da aplicação
func NewRouter(container *di.Container) http.Handler {
	r := chi.NewRouter()

	// 1. Middlewares Globais (Configurações do Framework Web)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// 2. Health Check (Simples)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 3. Registro de Rotas da API (V1)
	r.Route("/api/v1", func(r chi.Router) {
		container.UserHandler.RegisterRoutes(r)
		container.OrgHandler.RegisterRoutes(r)
		// container.StoreHandler.RegisterRoutes(r) // Em breve...
	})

	return r
}
