package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
)

type OrganizationRepository interface {
	// Escrita
	Create(ctx context.Context, org *entity.Organization) error
	Update(ctx context.Context, org *entity.Organization) error // <--- Novo

	// Leitura
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Organization, error) // Útil para checar duplicidade
}
