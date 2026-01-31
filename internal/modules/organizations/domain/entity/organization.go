package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/validator"
)

// 1. Tipagem Forte para Planos (Enums)
type OrganizationPlan string

const (
	PlanFree       OrganizationPlan = "free"       // Teste / Limitado
	PlanPro        OrganizationPlan = "pro"        // Pequenas Lojas
	PlanEnterprise OrganizationPlan = "enterprise" // Grandes Redes (Customizado)
)

// 2. Tipagem para Setores de Atuação (Smart Gondola Context)
type OrganizationSector string

const (
	SectorSupermarket OrganizationSector = "supermarket"
	SectorPharmacy    OrganizationSector = "pharmacy"
	SectorRetail      OrganizationSector = "retail"    // Varejo Geral
	SectorWarehouse   OrganizationSector = "warehouse" // Centro de Distribuição
	SectorOther       OrganizationSector = "other"
)

// OrganizationSettings define as regras e limites do plano (JSONB no banco)
type OrganizationSettings struct {
	MaxUsers   int `json:"max_users"`
	MaxDevices int `json:"max_devices"` // Quantidade de etiquetas/gateways
	// Futuro: MaxDashboards, AllowAPI, etc.
}

// Organization representa a empresa cliente (Tenant)
type Organization struct {
	ID       uuid.UUID          `json:"id"`
	Name     string             `json:"name"`
	Document string             `json:"document"` // CNPJ
	Slug     string             `json:"slug"`
	Plan     OrganizationPlan   `json:"plan"`
	Sector   OrganizationSector `json:"sector"`
	IsActive bool               `json:"is_active"`

	Settings OrganizationSettings `json:"settings"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewOrganization cria uma nova empresa com validações
func NewOrganization(name, document, slug string, sector OrganizationSector, plan OrganizationPlan) (*Organization, error) {
	if name == "" {
		return nil, errors.New("nome da organização é obrigatório")
	}
	if slug == "" {
		return nil, errors.New("slug é obrigatório")
	}

	// Valida se o setor é permitido
	if !isValidSector(sector) {
		return nil, errors.New("setor de atuação inválido")
	}

	// Validação de CNPJ (Remove pontuação antes de validar)
	cleanDoc := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(document, ".", ""), "-", ""), "/", "")
	if !validator.IsCNPJ(cleanDoc) {
		return nil, errors.New("CNPJ inválido")
	}

	// Define limites baseados no plano escolhido
	defaultSettings := OrganizationSettings{}

	switch plan {
	case PlanFree:
		defaultSettings = OrganizationSettings{MaxUsers: 2, MaxDevices: 10}
	case PlanPro:
		defaultSettings = OrganizationSettings{MaxUsers: 10, MaxDevices: 500}
	case PlanEnterprise:
		defaultSettings = OrganizationSettings{MaxUsers: 9999, MaxDevices: 99999}
	default:
		defaultSettings = OrganizationSettings{MaxUsers: 2, MaxDevices: 10} // Default seguro
	}

	return &Organization{
		ID:        uuid.New(),
		Name:      name,
		Document:  document,
		Slug:      strings.ToLower(slug),
		Plan:      plan,
		Sector:    sector,
		Settings:  defaultSettings, // Salva os limites iniciais
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// Update permite alterar dados cadastrais da empresa
func (o *Organization) Update(name, document string, sector OrganizationSector) error {
	if name == "" {
		return errors.New("nome não pode ser vazio")
	}
	if !isValidSector(sector) {
		return errors.New("setor inválido")
	}

	// Valida CNPJ se foi alterado
	cleanDoc := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(document, ".", ""), "-", ""), "/", "")
	if !validator.IsCNPJ(cleanDoc) {
		return errors.New("CNPJ inválido")
	}

	o.Name = name
	o.Document = cleanDoc
	o.Sector = sector
	o.UpdatedAt = time.Now()

	return nil
}

// ChangePlan altera o plano e atualiza os limites (Settings)
func (o *Organization) ChangePlan(newPlan OrganizationPlan) {
	o.Plan = newPlan

	// Recalcula os limites baseados no novo plano
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

// Auxiliar para validar setor
func isValidSector(s OrganizationSector) bool {
	switch s {
	case SectorSupermarket, SectorPharmacy, SectorRetail, SectorWarehouse, SectorOther:
		return true
	}
	return false
}

// Deactivate desativa a organização inteira
func (o *Organization) Deactivate() {
	o.IsActive = false
	o.UpdatedAt = time.Now()
}

// Activate reativa a organização
func (o *Organization) Activate() {
	o.IsActive = true
	o.UpdatedAt = time.Now()
}
