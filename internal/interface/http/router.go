package http

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/paulochiaradia/smart-gondola-backend/internal/di"
	customMiddleware "github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/middleware"
)

func NewRouter(container *di.Container) http.Handler {
	r := chi.NewRouter()

	// 1. Configuração de CORS Dinâmica (Baseada no Ambiente)
	// Em prod, defina ALLOWED_ORIGINS="https://meu-front.com.br"
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var origins []string

	if allowedOriginsStr == "" {
		// Modo Desenvolvimento: Aceita de qualquer lugar
		origins = []string{"*"}
	} else {
		// Modo Produção: Aceita apenas os domínios listados
		origins = strings.Split(allowedOriginsStr, ",")
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 2. Middlewares Globais (Logs e Prevenção de Crash)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 3. Segurança Extra: Limite de Payload (2MB)
	// Impede que enviem um JSON de 1GB e travem a memória do servidor
	r.Use(customMiddleware.LimitPayloadSize(2 * 1024 * 1024))

	r.Route("/api/v1", func(r chi.Router) {
		// ===========================
		// ROTAS PÚBLICAS (Sem Token)
		// ===========================
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		r.Post("/auth/login", container.UserHandler.Login)
		r.Post("/auth/register", container.UserHandler.Register)

		// ===========================
		// ROTAS PROTEGIDAS (Com Token)
		// ===========================
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.AuthMiddleware)

			// Rotas de Organização
			r.Get("/organizations/{id}", container.OrgHandler.GetByID)

			// Rotas de Lojas
			r.With(customMiddleware.RequireRole("admin", "tenant")).
				Post("/stores", container.StoreHandler.Create)

			r.With(customMiddleware.RequireRole("admin", "tenant", "manager")).
				Get("/organizations/{orgId}/stores", container.StoreHandler.ListByOrg)
		})
	})

	return r
}
