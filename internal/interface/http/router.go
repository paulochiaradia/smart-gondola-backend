package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/paulochiaradia/smart-gondola-backend/internal/di"
	customMiddleware "github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/middleware" // Alias para nosso middleware
)

func NewRouter(container *di.Container) http.Handler {
	r := chi.NewRouter()

	// 1. Middlewares Globais (Logs, CORS, Recover)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Em prod, restrinja isso!
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		// ===========================
		// ROTAS PÚBLICAS (Sem Token)
		// ===========================

		// Health Check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})

		// Autenticação
		r.Post("/auth/login", container.UserHandler.Login)
		r.Post("/auth/register", container.UserHandler.Register)

		// (Opcional) Webhook de pagamento, etc.

		// ===========================
		// ROTAS PROTEGIDAS (Com Token)
		// ===========================
		// ===========================
		// ROTAS PROTEGIDAS (Com Token)
		// ===========================

		r.Group(func(r chi.Router) {
			// 1. Aplica o AuthMiddleware para TODAS as rotas deste grupo
			r.Use(customMiddleware.AuthMiddleware)

			// Rotas de Organização
			r.Get("/organizations/{id}", container.OrgHandler.GetByID)

			// Rotas de Lojas
			// Protegendo a criação de loja: Apenas 'admin' e 'tenant' (Dono) podem criar novas lojas.
			// O 'manager' gerencia a loja, mas não cria novas. O 'operator' só opera.
			r.With(customMiddleware.RequireRole("admin", "tenant")).
				Post("/stores", container.StoreHandler.Create)

			// Listar lojas: 'admin', 'tenant' e 'manager' podem ver.
			r.With(customMiddleware.RequireRole("admin", "tenant", "manager")).
				Get("/organizations/{orgId}/stores", container.StoreHandler.ListByOrg)
		})

	})

	return r
}
