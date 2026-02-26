package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
)

// AddressInput facilita a entrada de dados do endereço
type AddressInput struct {
	Street     string `json:"street" validate:"required"`
	Number     string `json:"number" validate:"required"`
	Complement string `json:"complement"`
	District   string `json:"district" validate:"required"`
	City       string `json:"city" validate:"required"`
	State      string `json:"state" validate:"required,len=2"`
	ZipCode    string `json:"zip_code" validate:"required"`
}

// CreateStoreRequest entrada para criar loja
type CreateStoreRequest struct {
	OrganizationID uuid.UUID    `json:"organization_id" validate:"required"`
	Name           string       `json:"name" validate:"required"`
	Code           string       `json:"code" validate:"required"`
	Timezone       string       `json:"timezone"`
	Address        AddressInput `json:"address"`
}

// UpdateStoreRequest entrada para editar loja
type UpdateStoreRequest struct {
	Name     string       `json:"name"`
	Address  AddressInput `json:"address"`
	Timezone string       `json:"timezone"`
}

// StoreResponse saída completa
type StoreResponse struct {
	ID             uuid.UUID           `json:"id"`
	OrganizationID uuid.UUID           `json:"organization_id"`
	Name           string              `json:"name"`
	Code           string              `json:"code"`
	Address        entity.StoreAddress `json:"address"`
	Timezone       string              `json:"timezone"`
	IsActive       bool                `json:"is_active"`
	CreatedAt      time.Time           `json:"created_at"`
}
