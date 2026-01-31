package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/repository"
)

type OrganizationUseCase struct {
	repo repository.OrganizationRepository
}

func NewOrganizationUseCase(repo repository.OrganizationRepository) *OrganizationUseCase {
	return &OrganizationUseCase{repo: repo}
}

// Create cria uma nova organização (Geralmente chamado pelo SuperAdmin ou no Sign Up)
func (uc *OrganizationUseCase) Create(ctx context.Context, input dto.CreateOrganizationRequest) (*dto.OrganizationResponse, error) {
	// 1. Verifica se o slug já está em uso (Regra de Negócio)
	existingOrg, err := uc.repo.GetBySlug(ctx, input.Slug)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar slug: %w", err)
	}
	if existingOrg != nil {
		return nil, errors.New("este slug já está em uso por outra empresa")
	}

	// 2. Cria a entidade (Já valida CNPJ, Setor e define limites do plano)
	org, err := entity.NewOrganization(input.Name, input.Document, input.Slug, input.Sector, input.Plan)
	if err != nil {
		return nil, err
	}

	// 3. Persiste no banco
	if err := uc.repo.Create(ctx, org); err != nil {
		return nil, err
	}

	// 4. Retorna DTO
	return uc.toResponse(org), nil
}

// GetByID busca os dados da empresa (Usado pelo endpoint /me ou middleware)
func (uc *OrganizationUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.OrganizationResponse, error) {
	org, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, errors.New("organização não encontrada")
	}

	return uc.toResponse(org), nil
}

// Update atualiza dados cadastrais
func (uc *OrganizationUseCase) Update(ctx context.Context, id uuid.UUID, input dto.UpdateOrganizationRequest) (*dto.OrganizationResponse, error) {
	// 1. Busca atual
	org, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, errors.New("organização não encontrada")
	}

	// 2. Aplica as mudanças na entidade (valida novo CNPJ se mudou)
	if err := org.Update(input.Name, input.Document, input.Sector); err != nil {
		return nil, err
	}

	// 3. Salva
	if err := uc.repo.Update(ctx, org); err != nil {
		return nil, err
	}

	return uc.toResponse(org), nil
}

// Helper para converter Entity -> Response DTO
func (uc *OrganizationUseCase) toResponse(org *entity.Organization) *dto.OrganizationResponse {
	return &dto.OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Document:  org.Document,
		Slug:      org.Slug,
		Plan:      org.Plan,
		Sector:    org.Sector,
		Settings:  org.Settings,
		IsActive:  org.IsActive,
		CreatedAt: org.CreatedAt,
	}
}
