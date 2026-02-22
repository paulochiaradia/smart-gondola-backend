package middleware

import (
	"net/http"

	"github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/response"
)

// RequireRole é um middleware que bloqueia o acesso se o usuário não tiver uma das roles permitidas.
// Ele DEVE ser usado nas rotas logo após o AuthMiddleware.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Pegamos a role que o AuthMiddleware injetou no contexto
			userRole := GetRole(r.Context())

			// Verifica se a role do usuário está na lista de permitidas
			for _, role := range allowedRoles {
				if userRole == role {
					// Tem permissão! Passa para o próximo (o Handler da rota)
					next.ServeHTTP(w, r)
					return
				}
			}

			// Se chegou aqui, não tem permissão. Retorna 403 Forbidden.
			response.Error(w, http.StatusForbidden, "Acesso negado: seu perfil não tem permissão para esta ação")
		})
	}
}
