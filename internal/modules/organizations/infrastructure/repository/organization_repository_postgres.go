package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/repository"
)

type OrganizationRepoPostgres struct {
	db *sql.DB
}

// NewOrganizationRepository cria a instância
func NewOrganizationRepository(db *sql.DB) repository.OrganizationRepository {
	return &OrganizationRepoPostgres{db: db}
}

// Create insere uma nova organização
func (r *OrganizationRepoPostgres) Create(ctx context.Context, org *entity.Organization) error {
	// 1. Converte Settings (struct) para JSON (bytes)
	settingsJSON, err := json.Marshal(org.Settings)
	if err != nil {
		return fmt.Errorf("erro ao serializar settings: %w", err)
	}

	query := `
		INSERT INTO organizations (
			id, name, document, slug, plan, sector, settings, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		org.ID,
		org.Name,
		org.Document,
		org.Slug,
		org.Plan,
		org.Sector,
		settingsJSON, // Grava o JSON no banco
		org.IsActive,
		org.CreatedAt,
		org.UpdatedAt,
	)

	return err
}

// Update atualiza os dados da organização
func (r *OrganizationRepoPostgres) Update(ctx context.Context, org *entity.Organization) error {
	settingsJSON, err := json.Marshal(org.Settings)
	if err != nil {
		return fmt.Errorf("erro ao serializar settings no update: %w", err)
	}

	query := `
		UPDATE organizations SET
			name = $1,
			document = $2,
			sector = $3,
			plan = $4,
			settings = $5,
			is_active = $6,
			updated_at = $7
		WHERE id = $8
	`

	_, err = r.db.ExecContext(ctx, query,
		org.Name,
		org.Document,
		org.Sector,
		org.Plan,
		settingsJSON,
		org.IsActive,
		org.UpdatedAt, // Importante atualizar a data
		org.ID,
	)

	return err
}

// GetByID busca pelo ID
func (r *OrganizationRepoPostgres) GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error) {
	query := `
		SELECT 
			id, name, document, slug, plan, sector, settings, is_active, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`

	var org entity.Organization
	var settingsBytes []byte // Buffer temporário para o JSON

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Document,
		&org.Slug,
		&org.Plan,
		&org.Sector,
		&settingsBytes, // O banco joga os bytes aqui
		&org.IsActive,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Converte de volta: JSON (banco) -> Struct (Go)
	if len(settingsBytes) > 0 {
		if err := json.Unmarshal(settingsBytes, &org.Settings); err != nil {
			return nil, fmt.Errorf("erro ao desserializar settings: %w", err)
		}
	}

	return &org, nil
}

// GetBySlug busca pelo slug (útil para verificar duplicidade ou rota de API)
func (r *OrganizationRepoPostgres) GetBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	query := `
		SELECT 
			id, name, document, slug, plan, sector, settings, is_active, created_at, updated_at
		FROM organizations
		WHERE slug = $1
	`

	var org entity.Organization
	var settingsBytes []byte

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Document,
		&org.Slug,
		&org.Plan,
		&org.Sector,
		&settingsBytes,
		&org.IsActive,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if len(settingsBytes) > 0 {
		if err := json.Unmarshal(settingsBytes, &org.Settings); err != nil {
			return nil, fmt.Errorf("erro ao desserializar settings: %w", err)
		}
	}

	return &org, nil
}
