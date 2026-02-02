package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/repository"
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
	// Dica: Se o código for duplicado na mesma org, o banco vai retornar erro.
	// Poderíamos tratar o erro específico do driver aqui ou deixar o Handler pegar.
	if err := uc.repo.Create(ctx, store); err != nil {
		// Tratamento simples para erro de duplicidade
		if strings.Contains(err.Error(), "uq_stores_org_code") {
			return nil, errors.New("já existe uma loja com este código nesta organização")
		}
		return nil, err
	}

	return uc.toResponse(store), nil
}

func (uc *StoreUseCase) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*dto.StoreResponse, error) {
	stores, err := uc.repo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var response []*dto.StoreResponse
	for _, s := range stores {
		response = append(response, uc.toResponse(s))
	}
	return response, nil
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
