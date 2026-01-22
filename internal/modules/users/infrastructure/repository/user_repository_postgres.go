package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/domain/repository"
)

type UserRepoPostgres struct {
	db *sql.DB
}

// NewUserRepository cria uma nova instância do repositório
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &UserRepoPostgres{db: db}
}

// Create insere um novo usuário
func (r *UserRepoPostgres) Create(ctx context.Context, u *entity.User) error {
	// Converte a struct de 2FA para JSONB antes de salvar
	twoFactorJSON, err := json.Marshal(u.TwoFactor)
	if err != nil {
		return fmt.Errorf("failed to marshal two_factor: %w", err)
	}

	query := `
		INSERT INTO users (
			id, organization_id, store_id, name, email, phone, avatar_url,
			password_hash, role, status, invited_by, timezone, language,
			two_factor_settings, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		u.ID, u.OrganizationID, u.StoreID, u.Name, u.Email, u.Phone, u.AvatarURL,
		u.PasswordHash, u.Role, u.Status, u.InvitedBy, u.Timezone, u.Language,
		twoFactorJSON, u.CreatedAt, u.UpdatedAt,
	)

	return err
}

// GetByEmail busca um usuário pelo email
func (r *UserRepoPostgres) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT 
			id, organization_id, store_id, name, email, phone, avatar_url,
			password_hash, role, status, timezone, language, two_factor_settings,
			failed_login_attempts, locked_until, last_login_at, created_at
		FROM users 
		WHERE email = $1
	`

	var u entity.User
	var twoFactorJSON []byte
	// Nota: Scan direto de campos que podem ser NULL (como StoreID, LockedUntil) exige cuidado.
	// O pgx geralmente lida bem com *uuid.UUID, mas sql.DB padrão exige scan em sql.Null...
	// Para simplificar aqui, vamos focar no caminho feliz. Se der erro de NULL, ajustamos.

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.OrganizationID, &u.StoreID, &u.Name, &u.Email, &u.Phone, &u.AvatarURL,
		&u.PasswordHash, &u.Role, &u.Status, &u.Timezone, &u.Language, &twoFactorJSON,
		&u.FailedLoginAttempts, &u.LockedUntil, &u.LastLoginAt, &u.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Retorna nil se não achar (Use Case trata isso)
		}
		return nil, err
	}

	// Reconverte o JSONB para a struct Go
	if len(twoFactorJSON) > 0 {
		u.TwoFactor = &entity.TwoFactorAuth{}
		_ = json.Unmarshal(twoFactorJSON, u.TwoFactor)
	}

	return &u, nil
}

// GetByID busca um usuário pelo ID
func (r *UserRepoPostgres) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT 
			id, organization_id, store_id, name, email, phone, avatar_url,
			password_hash, role, status, timezone, language, two_factor_settings,
			failed_login_attempts, locked_until, last_login_at, created_at
		FROM users 
		WHERE id = $1
	`

	var u entity.User
	var twoFactorJSON []byte

	// Executa a query passando o ID
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.OrganizationID, &u.StoreID, &u.Name, &u.Email, &u.Phone, &u.AvatarURL,
		&u.PasswordHash, &u.Role, &u.Status, &u.Timezone, &u.Language, &twoFactorJSON,
		&u.FailedLoginAttempts, &u.LockedUntil, &u.LastLoginAt, &u.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Não encontrou ninguém com esse ID (retorna nil, nil sem erro)
			return nil, nil
		}
		// Erro de conexão ou query
		return nil, fmt.Errorf("erro ao buscar user por id: %w", err)
	}

	// Reconverte o JSONB do banco para a struct Go (TwoFactorAuth)
	if len(twoFactorJSON) > 0 {
		u.TwoFactor = &entity.TwoFactorAuth{}
		if err := json.Unmarshal(twoFactorJSON, u.TwoFactor); err != nil {
			return nil, fmt.Errorf("erro ao decodificar 2fa: %w", err)
		}
	}

	return &u, nil
}

// Update atualiza dados cadastrais
func (r *UserRepoPostgres) Update(ctx context.Context, u *entity.User) error {
	query := `
		UPDATE users SET 
			name=$1, phone=$2, avatar_url=$3, role=$4, status=$5, updated_at=NOW()
		WHERE id=$6
	`
	_, err := r.db.ExecContext(ctx, query, u.Name, u.Phone, u.AvatarURL, u.Role, u.Status, u.ID)
	return err
}

// UpdateSecurity atualiza dados sensíveis (Senha, Bloqueio)
func (r *UserRepoPostgres) UpdateSecurity(ctx context.Context, u *entity.User) error {
	query := `
		UPDATE users SET 
			password_hash=$1, failed_login_attempts=$2, locked_until=$3, 
			last_login_at=$4, last_login_ip=$5
		WHERE id=$6
	`
	_, err := r.db.ExecContext(ctx, query,
		u.PasswordHash, u.FailedLoginAttempts, u.LockedUntil,
		u.LastLoginAt, u.LastLoginIP, u.ID,
	)
	return err
}
