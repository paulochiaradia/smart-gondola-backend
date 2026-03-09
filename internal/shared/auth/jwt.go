package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken cria um JWT com os dados do usuário E da organização
func GenerateToken(userID uuid.UUID, orgID uuid.UUID, role string) (string, error) {
	cfg := config.Get()

	// Adiciona as Claims (As informações que vão dentro do envelope)
	claims := jwt.MapClaims{
		"user_id": userID.String(), // Ajustado para "user_id"
		"org_id":  orgID.String(),  // <-- ADICIONAMOS A ORGANIZAÇÃO AQUI!
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expira em 24h
		"iat":     time.Now().Unix(),
		"iss":     "smart-gondola-api", // Nome correto da sua API
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.Get()
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token inválido")
	}
	return claims, nil
}
