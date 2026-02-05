package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
)

// Chaves para salvar/recuperar dados do contexto
type contextKey string

const (
	UserContextKey = contextKey("user_id")
	OrgContextKey  = contextKey("org_id")
	RoleContextKey = contextKey("role")
)

// AuthMiddleware valida o JWT e injeta user/org no contexto
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Pegar o Header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		// 2. Remover o prefixo "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // Não tinha "Bearer "
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// 3. Parse e Validação do Token
		cfg := config.Get()
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Valida o algoritmo de assinatura (segurança crítica!)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// 4. Extrair Claims (Dados do Payload)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// 5. Converter IDs para UUID e Injetar no Contexto
		// Assumindo que o token tem campos "sub" (user_id) e "org_id"
		userIDStr, _ := claims["sub"].(string)
		orgIDStr, _ := claims["org_id"].(string)
		role, _ := claims["role"].(string)

		userID, errU := uuid.Parse(userIDStr)
		orgID, errO := uuid.Parse(orgIDStr)

		if errU != nil || errO != nil {
			http.Error(w, "Token does not contain valid IDs", http.StatusUnauthorized)
			return
		}

		// Injeta no contexto
		ctx := context.WithValue(r.Context(), UserContextKey, userID)
		ctx = context.WithValue(ctx, OrgContextKey, orgID)
		ctx = context.WithValue(ctx, RoleContextKey, role)

		// Passa para o próximo handler com o contexto enriquecido
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helpers para recuperar dados do contexto nos Handlers
func GetUserID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(UserContextKey).(uuid.UUID)
	return id
}

func GetOrgID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(OrgContextKey).(uuid.UUID)
	return id
}

func GetRole(ctx context.Context) string {
	role, _ := ctx.Value(RoleContextKey).(string)
	return role
}
