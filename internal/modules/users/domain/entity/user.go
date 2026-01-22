package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRole define os papéis de acesso (RBAC)
type UserRole string

const (
	RoleAdmin    UserRole = "admin"    // Admin do SaaS (Você)
	RoleTenant   UserRole = "tenant"   // Dono da Empresa Cliente
	RoleManager  UserRole = "manager"  // Gerente de Loja
	RoleOperator UserRole = "operator" // Operador de Loja
)

// UserStatus define o ciclo de vida do cadastro
type UserStatus string

const (
	StatusPending   UserStatus = "pending"   // Aguardando aprovação
	StatusActive    UserStatus = "active"    // Ativo e operante
	StatusSuspended UserStatus = "suspended" // Bloqueado (falta de pagamento/demissão)
)

// User representa a identidade central
type User struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`    // Multi-tenancy (Obrigatório)
	StoreID        *uuid.UUID `json:"store_id,omitempty"` // Opcional: Restringe a uma loja

	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`      // Mobile First
	AvatarURL string `json:"avatar_url,omitempty"` // UX Mobile

	PasswordHash string     `json:"-"`
	Role         UserRole   `json:"role"`
	Status       UserStatus `json:"status"`

	// Rastreabilidade de Cadastro
	InvitedBy *uuid.UUID `json:"invited_by,omitempty"`

	// Preferências (I18n & L10n)
	Timezone string `json:"timezone"` // Ex: "America/Sao_Paulo"
	Language string `json:"language"` // Ex: "pt-BR"

	// Segurança & Compliance
	EmailVerifiedAt   *time.Time     `json:"email_verified_at,omitempty"`
	TermsAcceptedAt   *time.Time     `json:"terms_accepted_at,omitempty"`
	PasswordChangedAt *time.Time     `json:"password_changed_at,omitempty"`
	TwoFactor         *TwoFactorAuth `json:"two_factor,omitempty"`

	// Auditoria
	FailedLoginAttempts int        `json:"-"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP         string     `json:"last_login_ip,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TwoFactorAuth Configurações de 2FA (Value Object)
type TwoFactorAuth struct {
	Enabled   bool   `json:"enabled"`
	Secret    string `json:"-"`
	QRCodeURL string `json:"qr_code_url,omitempty"`
}

// NewUser Factory - Prepara usuário para fluxo de auto-cadastro ou convite
func NewUser(orgID uuid.UUID, name, email, password string, role UserRole) (*User, error) {
	if orgID == uuid.Nil {
		return nil, errors.New("organization_id é obrigatório")
	}
	if name == "" || email == "" {
		return nil, errors.New("nome e email são obrigatórios")
	}
	if len(password) < 6 {
		return nil, errors.New("senha deve ter no mínimo 6 caracteres")
	}

	u := &User{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Email:          email,
		Role:           role,
		Status:         StatusActive, // Default ativo, mude para Pending se quiser aprovação
		Timezone:       "UTC",        // Default seguro
		Language:       "pt-BR",      // Default seguro
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		TwoFactor:      &TwoFactorAuth{Enabled: false},
	}

	if err := u.SetPassword(password); err != nil {
		return nil, err
	}

	return u, nil
}

// SetPassword encripta a senha e registra a data da mudança
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	now := time.Now()
	u.PasswordChangedAt = &now
	return nil
}

// CheckPassword valida a senha
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// IsLocked verifica bloqueio temporário
func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && u.LockedUntil.After(time.Now())
}
