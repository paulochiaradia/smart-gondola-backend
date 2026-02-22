package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/repository"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/pagination" // Importe o pacote de paginação
)

type StoreUseCase struct {
	repo repository.StoreRepository
}

func NewStoreUseCase(repo repository.StoreRepository) *StoreUseCase {
	return &StoreUseCase{repo: repo}
}

func (uc *StoreUseCase) Create(ctx context.Context, input dto.CreateStoreRequest) (*dto.StoreResponse, error) {
	// 1. Cria a Entidade
	store, err := entity.NewStore(input.OrganizationID, input.Name, input.Code, input.Timezone)
	if err != nil {
		return nil, err
	}

	// 2. Preenche o Endereço (Mapeamento DTO -> Entity)
	store.UpdateAddress(entity.StoreAddress{
		Street:     input.Address.Street,
		Number:     input.Address.Number,
		Complement: input.Address.Complement,
		District:   input.Address.District,
		City:       input.Address.City,
		State:      input.Address.State,
		ZipCode:    input.Address.ZipCode,
	})

	// 3. Persiste no Banco
	if err := uc.repo.Create(ctx, store); err != nil {
		if strings.Contains(err.Error(), "uq_stores_org_code") {
			return nil, errors.New("já existe uma loja com este código nesta organização")
		}
		return nil, err
	}

	return uc.toResponse(store), nil
}

// ListByOrganization recebe os parâmetros, manda o Repositório buscar no banco, e converte para DTO
func (uc *StoreUseCase) ListByOrganization(ctx context.Context, orgID uuid.UUID, params pagination.Params) ([]*dto.StoreResponse, int64, error) {
	stores, totalItems, err := uc.repo.ListByOrganization(ctx, orgID, params)
	if err != nil {
		return nil, 0, err
	}

	// 2. Converte as Entidades para DTOs de Resposta
	var response []*dto.StoreResponse
	for _, s := range stores {
		response = append(response, uc.toResponse(s))
	}

	return response, totalItems, nil
}

func (uc *StoreUseCase) GetByID(ctx context.Context, id uuid.UUID) (*dto.StoreResponse, error) {
	store, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("loja não encontrada")
	}
	return uc.toResponse(store), nil
}

func (uc *StoreUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

// Helper de Conversão Entity -> DTO Response
func (uc *StoreUseCase) toResponse(s *entity.Store) *dto.StoreResponse {
	return &dto.StoreResponse{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		Name:           s.Name,
		Code:           s.Code,
		Timezone:       s.Timezone,
		IsActive:       s.IsActive,
		CreatedAt:      s.CreatedAt,
		Address:        s.Address,
	}
}
