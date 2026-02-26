package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
)

// CreateOrganizationRequest é o que esperamos receber no cadastro
type CreateOrganizationRequest struct {
	Name     string                    `json:"name" validate:"required"`
	Document string                    `json:"document" validate:"required"` // CNPJ
	Slug     string                    `json:"slug" validate:"required"`
	Sector   entity.OrganizationSector `json:"sector" validate:"required,oneof=supermarket pharmacy retail warehouse other"`
	Plan     entity.OrganizationPlan   `json:"plan" validate:"required,oneof=free pro enterprise"`
}

// UpdateOrganizationRequest para atualizações cadastrais
type UpdateOrganizationRequest struct {
	Name     string                    `json:"name" validate:"required"`
	Document string                    `json:"document" validate:"required"`
	Sector   entity.OrganizationSector `json:"sector" validate:"required,oneof=supermarket pharmacy retail warehouse other"`
}

// OrganizationResponse é o que devolvemos para o frontend
type OrganizationResponse struct {
	ID        uuid.UUID                   `json:"id"`
	Name      string                      `json:"name"`
	Document  string                      `json:"document"`
	Slug      string                      `json:"slug"`
	Plan      entity.OrganizationPlan     `json:"plan"`
	Sector    entity.OrganizationSector   `json:"sector"`
	Settings  entity.OrganizationSettings `json:"settings"`
	IsActive  bool                        `json:"is_active"`
	CreatedAt time.Time                   `json:"created_at"`
}
