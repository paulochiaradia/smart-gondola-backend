package entity

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/validator"
)

// Tipos Enum
type OrganizationPlan string

const (
	PlanFree       OrganizationPlan = "free"
	PlanPro        OrganizationPlan = "pro"
	PlanEnterprise OrganizationPlan = "enterprise"
)

type OrganizationSector string

const (
	SectorSupermarket OrganizationSector = "supermarket"
	SectorPharmacy    OrganizationSector = "pharmacy"
	SectorRetail      OrganizationSector = "retail"
	SectorWarehouse   OrganizationSector = "warehouse"
	SectorOther       OrganizationSector = "other"
)

type OrganizationSettings struct {
	MaxUsers   int `json:"max_users"`
	MaxDevices int `json:"max_devices"`
}

type Organization struct {
	ID        uuid.UUID            `json:"id"`
	Name      string               `json:"name"`
	Document  string               `json:"document"` // Será salvo sempre LIMPO (apenas números)
	Slug      string               `json:"slug"`
	Plan      OrganizationPlan     `json:"plan"`
	Sector    OrganizationSector   `json:"sector"`
	IsActive  bool                 `json:"is_active"`
	Settings  OrganizationSettings `json:"settings"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

func NewOrganization(name, document, slug string, sector OrganizationSector, plan OrganizationPlan) (*Organization, error) {
	if name == "" {
		return nil, errors.New("nome da organização é obrigatório")
	}
	if slug == "" {
		return nil, errors.New("slug é obrigatório")
	}

	if !isValidSector(sector) {
		return nil, errors.New("setor de atuação inválido")
	}

	// 1. Limpeza do CNPJ (Remove tudo que não for dígito)
	// Isso garante que "12.345..." vire "12345..."
	reg, _ := regexp.Compile("[^0-9]+")
	cleanDoc := reg.ReplaceAllString(document, "")

	// 2. Validação
	if !validator.IsCNPJ(cleanDoc) {
		return nil, errors.New("CNPJ inválido")
	}

	// 3. Define Limites
	defaultSettings := OrganizationSettings{}
	switch plan {
	case PlanFree:
		defaultSettings = OrganizationSettings{MaxUsers: 2, MaxDevices: 10}
	case PlanPro:
		defaultSettings = OrganizationSettings{MaxUsers: 10, MaxDevices: 500}
	case PlanEnterprise:
		defaultSettings = OrganizationSettings{MaxUsers: 9999, MaxDevices: 99999}
	default:
		defaultSettings = OrganizationSettings{MaxUsers: 2, MaxDevices: 10}
	}

	return &Organization{
		ID:        uuid.New(),
		Name:      name,
		Document:  cleanDoc, // <--- CORREÇÃO: Salva o limpo no banco
		Slug:      strings.ToLower(slug),
		Plan:      plan,
		Sector:    sector,
		Settings:  defaultSettings,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (o *Organization) Update(name, document string, sector OrganizationSector) error {
	if name == "" {
		return errors.New("nome não pode ser vazio")
	}
	if !isValidSector(sector) {
		return errors.New("setor inválido")
	}

	reg, _ := regexp.Compile("[^0-9]+")
	cleanDoc := reg.ReplaceAllString(document, "")

	if !validator.IsCNPJ(cleanDoc) {
		return errors.New("CNPJ inválido")
	}

	o.Name = name
	o.Document = cleanDoc
	o.Sector = sector
	o.UpdatedAt = time.Now()

	return nil
}

func (o *Organization) ChangePlan(newPlan OrganizationPlan) {
	o.Plan = newPlan
	switch newPlan {
	case PlanFree:
		o.Settings = OrganizationSettings{MaxUsers: 2, MaxDevices: 10}
	case PlanPro:
		o.Settings = OrganizationSettings{MaxUsers: 10, MaxDevices: 500}
	case PlanEnterprise:
		o.Settings = OrganizationSettings{MaxUsers: 9999, MaxDevices: 99999}
	}
	o.UpdatedAt = time.Now()
}

func isValidSector(s OrganizationSector) bool {
	switch s {
	case SectorSupermarket, SectorPharmacy, SectorRetail, SectorWarehouse, SectorOther:
		return true
	}
	return false
}

func (o *Organization) Deactivate() {
	o.IsActive = false
	o.UpdatedAt = time.Now()
}

func (o *Organization) Activate() {
	o.IsActive = true
	o.UpdatedAt = time.Now()
}
