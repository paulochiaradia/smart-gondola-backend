package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/response"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
)

// Chaves para salvar/recuperar dados do contexto
type contextKey string

const (
	UserContextKey = contextKey("user_id")
	OrgContextKey  = contextKey("org_id")
	RoleContextKey = contextKey("role")
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "Header 'Authorization' é obrigatório")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			response.Error(w, http.StatusUnauthorized, "Formato do token inválido. Use 'Bearer <token>'")
			return
		}

		cfg := config.Get()
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			response.Error(w, http.StatusUnauthorized, "Token inválido ou expirado")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "Payload do token inválido")
			return
		}

		userIDStr, _ := claims["sub"].(string)
		orgIDStr, _ := claims["org_id"].(string)
		role, _ := claims["role"].(string)

		userID, errU := uuid.Parse(userIDStr)
		orgID, errO := uuid.Parse(orgIDStr)

		if errU != nil || errO != nil {
			response.Error(w, http.StatusUnauthorized, "Token não contém IDs válidos")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userID)
		ctx = context.WithValue(ctx, OrgContextKey, orgID)
		ctx = context.WithValue(ctx, RoleContextKey, role)

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
