package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/pagination"
)

type StoreRepository interface {
	// Escrita
	Create(ctx context.Context, store *entity.Store) error
	Update(ctx context.Context, store *entity.Store) error
	Delete(ctx context.Context, id uuid.UUID) error // Soft Delete (Update deleted_at)

	// Leitura
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Store, error)
	ListByOrganization(ctx context.Context, orgID uuid.UUID, params pagination.Params) ([]*entity.Store, int64, error)
}
