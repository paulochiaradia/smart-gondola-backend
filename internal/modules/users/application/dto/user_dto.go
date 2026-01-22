package dto

import (
	"time"

	"github.com/google/uuid"
	// IMPORTANTE: Ajuste o caminho abaixo para o nome do seu módulo no go.mod
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/domain/entity"
)

// --- Auth DTOs (Login/Refresh) ---

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`

	// Metadados do App Mobile (Opcionais, mas vitais para UserDevice)
	DeviceID   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
	PushToken  string `json:"push_token,omitempty"`
	Platform   string `json:"platform,omitempty"`
	AppVersion string `json:"app_version,omitempty"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"`
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"` // Retorna dados básicos e avatar
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// --- User Management DTOs (CRUD) ---

type CreateUserRequest struct {
	// OrganizationID vem do token do admin ou do código de convite no Service
	OrganizationID uuid.UUID  `json:"organization_id" validate:"required"`
	StoreID        *uuid.UUID `json:"store_id"` // Opcional

	Name     string          `json:"name" validate:"required"`
	Email    string          `json:"email" validate:"required,email"`
	Phone    string          `json:"phone"`
	Password string          `json:"password" validate:"required,min=6"`
	Role     entity.UserRole `json:"role" validate:"oneof=tenant manager operator"`

	Timezone string `json:"timezone"`
	Language string `json:"language"`
}

type UpdateUserRequest struct {
	ID        uuid.UUID          `json:"id"`
	Name      string             `json:"name"`
	Phone     string             `json:"phone"`
	AvatarURL string             `json:"avatar_url"`
	Role      entity.UserRole    `json:"role"`
	Status    *entity.UserStatus `json:"status"` // Ponteiro para permitir mudança de status
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password" validate:"min=6"`
}

// --- Responses ---

type UserResponse struct {
	ID             uuid.UUID         `json:"id"`
	OrganizationID uuid.UUID         `json:"organization_id"`
	StoreID        *uuid.UUID        `json:"store_id,omitempty"`
	Name           string            `json:"name"`
	Email          string            `json:"email"`
	Phone          string            `json:"phone,omitempty"`
	AvatarURL      string            `json:"avatar_url,omitempty"`
	Role           entity.UserRole   `json:"role"`
	Status         entity.UserStatus `json:"status"`
	Timezone       string            `json:"timezone"`
	Language       string            `json:"language"`
	LastLogin      *time.Time        `json:"last_login,omitempty"`
}

// AuthLogResponse para histórico de segurança
type RecentLoginResponse struct {
	Date   time.Time `json:"date"`
	IP     string    `json:"ip"`
	Device string    `json:"device"` // Ex: "iPhone 13"
	Status string    `json:"status"` // Sucesso/Falha
}
