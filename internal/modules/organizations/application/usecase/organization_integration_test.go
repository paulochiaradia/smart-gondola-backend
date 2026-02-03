package usecase_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/usecase"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/infrastructure/repository"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/database"
)

type OrganizationSuite struct {
	suite.Suite
	db      *sql.DB
	useCase *usecase.OrganizationUseCase
}

func (s *OrganizationSuite) SetupSuite() {
	cfg := config.Get()
	if cfg.DBHost == "localhost" {
		cfg.DBHost = "127.0.0.1"
	}

	db, err := database.NewPostgres(cfg)
	s.Require().NoError(err)
	s.db = db

	repo := repository.NewOrganizationRepository(db)
	s.useCase = usecase.NewOrganizationUseCase(repo)
}

func (s *OrganizationSuite) SetupTest() {
	_, err := s.db.Exec("TRUNCATE organizations CASCADE")
	s.Require().NoError(err)
}

func (s *OrganizationSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

// --- TESTES ---

func (s *OrganizationSuite) TestCreateOrganization_Success() {
	// CNPJ REAL (Google Brasil): 06.990.590/0001-23
	// Enviamos limpo: 06990590000123
	input := dto.CreateOrganizationRequest{
		Name:     "Smart Gondola Matriz",
		Document: "06990590000123",
		Slug:     "smart-gondola-hq",
		Sector:   entity.SectorRetail,
		Plan:     entity.PlanEnterprise,
	}

	res, err := s.useCase.Create(context.Background(), input)

	if !s.NoError(err) {
		return
	}

	s.NotNil(res)
	s.NotEmpty(res.ID)
	s.Equal(input.Name, res.Name)
	s.Equal("06990590000123", res.Document)
	s.Equal(entity.PlanEnterprise, res.Plan)
}

func (s *OrganizationSuite) TestCreateOrganization_InvalidCNPJ() {
	input := dto.CreateOrganizationRequest{
		Name:     "Empresa Fake",
		Document: "00000000000000", // Zeros são inválidos
		Slug:     "empresa-fake",
		Sector:   entity.SectorOther,
		Plan:     entity.PlanFree,
	}

	res, err := s.useCase.Create(context.Background(), input)

	s.Error(err)
	s.Nil(res)
	s.Equal("CNPJ inválido", err.Error())
}

func (s *OrganizationSuite) TestCreateOrganization_DuplicateSlug() {
	// 1. Cria a primeira empresa (Facebook Brasil): 13.347.016/0001-17
	input1 := dto.CreateOrganizationRequest{
		Name:     "Loja A",
		Document: "13347016000117",
		Slug:     "minha-loja",
		Sector:   entity.SectorPharmacy,
		Plan:     entity.PlanFree,
	}
	_, err := s.useCase.Create(context.Background(), input1)
	s.NoError(err)

	// 2. Tenta criar a segunda com o MESMO slug
	// (Amazon Brasil): 15.436.940/0001-03
	input2 := dto.CreateOrganizationRequest{
		Name:     "Loja B",
		Document: "15436940000103",
		Slug:     "minha-loja", // Duplicado!
		Sector:   entity.SectorPharmacy,
		Plan:     entity.PlanFree,
	}
	res, err := s.useCase.Create(context.Background(), input2)

	s.Error(err)
	s.Nil(res)
	s.Contains(err.Error(), "slug") // Agora sim vai dar erro de slug, pois o CNPJ é válido
}

func TestOrganizationSuite(t *testing.T) {
	suite.Run(t, new(OrganizationSuite))
}
