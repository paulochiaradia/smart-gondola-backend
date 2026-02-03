package usecase_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/usecase"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/infrastructure/repository"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/database"
)

type StoreSuite struct {
	suite.Suite
	db           *sql.DB
	orgUseCase   *usecase.OrganizationUseCase
	storeUseCase *usecase.StoreUseCase

	// Vamos precisar de uma organização criada para pendurar as lojas
	defaultOrgID uuid.UUID
}

func (s *StoreSuite) SetupSuite() {
	cfg := config.Get()
	if cfg.DBHost == "localhost" {
		cfg.DBHost = "127.0.0.1"
	}

	db, err := database.NewPostgres(cfg)
	s.Require().NoError(err)

	// Inicializa Repos e UseCases
	orgRepo := repository.NewOrganizationRepository(db)
	s.orgUseCase = usecase.NewOrganizationUseCase(orgRepo)

	storeRepo := repository.NewStoreRepository(db)
	s.storeUseCase = usecase.NewStoreUseCase(storeRepo)

	s.db = db
}

func (s *StoreSuite) SetupTest() {
	// Limpa tudo (Cascade apaga as lojas também)
	_, err := s.db.Exec("TRUNCATE organizations CASCADE")
	s.Require().NoError(err)

	// CRIA UMA ORGANIZAÇÃO PADRÃO PARA OS TESTES
	// Usando CNPJ real (Magazine Luiza): 47.960.950/0001-21
	org, err := s.orgUseCase.Create(context.Background(), dto.CreateOrganizationRequest{
		Name:     "Magazine Teste",
		Document: "47960950000121",
		Slug:     "magalu-teste",
		Sector:   entity.SectorRetail,
		Plan:     entity.PlanPro,
	})
	s.Require().NoError(err)
	s.defaultOrgID = org.ID
}

func (s *StoreSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

// --- TESTES DE LOJA ---

func (s *StoreSuite) TestCreateStore_Success() {
	input := dto.CreateStoreRequest{
		OrganizationID: s.defaultOrgID,
		Name:           "Filial Centro",
		Code:           "LJ-001",
		Timezone:       "America/Sao_Paulo",
		Address: dto.AddressInput{
			Street:  "Rua das Palmeiras",
			Number:  "100",
			City:    "São Paulo",
			State:   "SP",
			ZipCode: "01000-000",
		},
	}

	res, err := s.storeUseCase.Create(context.Background(), input)

	if !s.NoError(err) {
		return
	}

	s.NotNil(res)
	s.Equal(input.Name, res.Name)
	s.Equal(input.Code, res.Code)
	s.Equal(input.Address.City, res.Address.City)
	s.Equal(s.defaultOrgID, res.OrganizationID)
}

func (s *StoreSuite) TestCreateStore_DuplicateCodeInSameOrg() {
	// 1. Cria Loja 1
	input1 := dto.CreateStoreRequest{
		OrganizationID: s.defaultOrgID,
		Name:           "Loja Um",
		Code:           "LJ-X", // Código X
		Timezone:       "UTC",
		Address:        dto.AddressInput{Street: "Rua A"},
	}
	_, err := s.storeUseCase.Create(context.Background(), input1)
	s.NoError(err)

	// 2. Tenta criar Loja 2 com MESMO código na MESMA org
	input2 := dto.CreateStoreRequest{
		OrganizationID: s.defaultOrgID,
		Name:           "Loja Dois",
		Code:           "LJ-X", // Duplicado!
		Timezone:       "UTC",
		Address:        dto.AddressInput{Street: "Rua B"},
	}
	res, err := s.storeUseCase.Create(context.Background(), input2)

	// Deve falhar
	s.Error(err)
	s.Nil(res)
	s.Contains(err.Error(), "código") // Esperamos erro de "código duplicado"
}

func (s *StoreSuite) TestCreateStore_SameCodeDifferentOrg() {
	// 1. Cria OUTRA Organização (Carrefour): 45.543.915/0001-81
	org2, err := s.orgUseCase.Create(context.Background(), dto.CreateOrganizationRequest{
		Name:     "Outro Mercado",
		Document: "45543915000181",
		Slug:     "outro-mercado",
		Sector:   entity.SectorSupermarket,
		Plan:     entity.PlanFree,
	})
	s.Require().NoError(err)

	// 2. Cria Loja na Org 1 com código "MATRIZ"
	_, err = s.storeUseCase.Create(context.Background(), dto.CreateStoreRequest{
		OrganizationID: s.defaultOrgID,
		Name:           "Minha Matriz",
		Code:           "MATRIZ",
		Address:        dto.AddressInput{Street: "Rua 1"},
	})
	s.NoError(err)

	// 3. Cria Loja na Org 2 TAMBÉM com código "MATRIZ"
	// ISSO DEVE FUNCIONAR (Multi-tenant: códigos iguais em orgs diferentes é permitido)
	res, err := s.storeUseCase.Create(context.Background(), dto.CreateStoreRequest{
		OrganizationID: org2.ID,
		Name:           "Matriz Deles",
		Code:           "MATRIZ",
		Address:        dto.AddressInput{Street: "Rua 2"},
	})

	s.NoError(err)
	s.NotNil(res)
	s.Equal("MATRIZ", res.Code)
}

func TestStoreSuite(t *testing.T) {
	suite.Run(t, new(StoreSuite))
}
