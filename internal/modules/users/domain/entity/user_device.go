package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserDevice representa um aparelho celular/tablet logado no App
// Essencial para enviar Push Notifications de alertas da gôndola
type UserDevice struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`

	// Identificação do Hardware
	DeviceID   string `json:"device_id"`   // ID único do hardware (Android ID / UUID iOS)
	DeviceName string `json:"device_name"` // Ex: "iPhone 13 de João"
	Platform   string `json:"platform"`    // "ios", "android", "web"
	AppVersion string `json:"app_version"` // Ex: "1.0.2" - Útil para forçar update

	// Notificações Push (FCM / APNS)
	PushToken string `json:"push_token"`

	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewUserDevice cria um registro de dispositivo
func NewUserDevice(userID uuid.UUID, deviceID, name, platform, pushToken, version string) *UserDevice {
	return &UserDevice{
		ID:           uuid.New(),
		UserID:       userID,
		DeviceID:     deviceID,
		DeviceName:   name,
		Platform:     platform,
		AppVersion:   version,
		PushToken:    pushToken,
		LastActiveAt: time.Now(),
		CreatedAt:    time.Now(),
	}
}

// UpdatePushToken atualiza o token caso ele expire (comum no Firebase) e a versão do app
func (d *UserDevice) UpdatePushToken(newToken, appVersion string) {
	d.PushToken = newToken
	d.AppVersion = appVersion
	d.LastActiveAt = time.Now()
}
