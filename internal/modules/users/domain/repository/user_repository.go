package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/domain/entity"
)

// UserRepository define as operações de banco de dados para Usuários
type UserRepository interface {
	// Comandos (Escrita)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	UpdateSecurity(ctx context.Context, user *entity.User) error // Apenas senha, bloqueios, etc.

	// Consultas (Leitura)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}
