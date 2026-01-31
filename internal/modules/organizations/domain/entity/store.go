package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// StoreAddress define onde a loja fica (Value Object)
type StoreAddress struct {
	Street     string `json:"street"`
	Number     string `json:"number"`
	Complement string `json:"complement"`
	District   string `json:"district"` // Bairro
	City       string `json:"city"`
	State      string `json:"state"`    // UF (SP, RJ, etc)
	ZipCode    string `json:"zip_code"` // CEP
}

// Store representa uma unidade física (Filial)
type Store struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`

	Name string `json:"name"`
	Code string `json:"code"` // Ex: LJ-001

	// Configurações Locais
	Address  StoreAddress `json:"address"`
	Timezone string       `json:"timezone"` // Ex: "America/Sao_Paulo"

	IsActive  bool       `json:"is_active"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // Ponteiro para suportar NULL (Soft Delete)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewStore cria uma nova filial
func NewStore(orgID uuid.UUID, name, code, timezone string) (*Store, error) {
	if name == "" {
		return nil, errors.New("nome da loja é obrigatório")
	}
	if orgID == uuid.Nil {
		return nil, errors.New("loja deve pertencer a uma organização")
	}
	// Default timezone se vier vazio
	if timezone == "" {
		timezone = "America/Sao_Paulo"
	}

	return &Store{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Code:           code,
		Timezone:       timezone,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

// UpdateAddress atualiza o endereço (Isolamento de lógica)
func (s *Store) UpdateAddress(addr StoreAddress) {
	s.Address = addr
	s.UpdatedAt = time.Now()
}

// Deactivate desativa a loja (Soft Delete ou Bloqueio)
func (s *Store) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now()
}
