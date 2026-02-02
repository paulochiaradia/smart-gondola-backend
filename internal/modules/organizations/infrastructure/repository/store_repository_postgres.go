package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/repository"
)

type StoreRepoPostgres struct {
	db *sql.DB
}

func NewStoreRepository(db *sql.DB) repository.StoreRepository {
	return &StoreRepoPostgres{db: db}
}

func (r *StoreRepoPostgres) Create(ctx context.Context, s *entity.Store) error {
	query := `
		INSERT INTO stores (
			id, organization_id, name, code, timezone, is_active,
			address_street, address_number, address_complement, address_district, 
			address_city, address_state, address_zip_code,
			created_at, updated_at, deleted_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, 
			$7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.OrganizationID, s.Name, s.Code, s.Timezone, s.IsActive,
		// Mapeando a struct Address para as colunas
		s.Address.Street, s.Address.Number, s.Address.Complement, s.Address.District,
		s.Address.City, s.Address.State, s.Address.ZipCode,
		s.CreatedAt, s.UpdatedAt, s.DeletedAt,
	)
	return err
}

func (r *StoreRepoPostgres) Update(ctx context.Context, s *entity.Store) error {
	query := `
		UPDATE stores SET
			name = $1, 
			timezone = $2,
			address_street = $3, address_number = $4, address_complement = $5,
			address_district = $6, address_city = $7, address_state = $8, address_zip_code = $9,
			updated_at = $10
		WHERE id = $11 AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query,
		s.Name, s.Timezone,
		s.Address.Street, s.Address.Number, s.Address.Complement,
		s.Address.District, s.Address.City, s.Address.State, s.Address.ZipCode,
		s.UpdatedAt, s.ID,
	)
	return err
}

func (r *StoreRepoPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	// Soft Delete: Apenas preenche o deleted_at
	query := `UPDATE stores SET deleted_at = $1, is_active = false WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *StoreRepoPostgres) GetByID(ctx context.Context, id uuid.UUID) (*entity.Store, error) {
	query := `
		SELECT 
			id, organization_id, name, code, timezone, is_active,
			address_street, address_number, address_complement, address_district, 
			address_city, address_state, address_zip_code,
			created_at, updated_at
		FROM stores 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var s entity.Store
	// Precisamos de variaveis auxiliares para NullStrings se o endereço for opcional,
	// mas como definimos colunas normais, vamos assumir string vazia se for null.

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.OrganizationID, &s.Name, &s.Code, &s.Timezone, &s.IsActive,
		&s.Address.Street, &s.Address.Number, &s.Address.Complement, &s.Address.District,
		&s.Address.City, &s.Address.State, &s.Address.ZipCode,
		&s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *StoreRepoPostgres) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*entity.Store, error) {
	query := `
		SELECT 
			id, organization_id, name, code, timezone, is_active,
			address_street, address_number, address_complement, address_district, 
			address_city, address_state, address_zip_code,
			created_at, updated_at
		FROM stores 
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []*entity.Store
	for rows.Next() {
		var s entity.Store
		if err := rows.Scan(
			&s.ID, &s.OrganizationID, &s.Name, &s.Code, &s.Timezone, &s.IsActive,
			&s.Address.Street, &s.Address.Number, &s.Address.Complement, &s.Address.District,
			&s.Address.City, &s.Address.State, &s.Address.ZipCode,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		stores = append(stores, &s)
	}
	return stores, nil
}
