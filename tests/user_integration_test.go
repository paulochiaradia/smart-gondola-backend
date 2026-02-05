package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/paulochiaradia/smart-gondola-backend/internal/di"
	routerLib "github.com/paulochiaradia/smart-gondola-backend/internal/interface/http"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	userDTO "github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/database"
)

type UserE2ESuite struct {
	suite.Suite
	db        *sql.DB
	handler   http.Handler
	container *di.Container

	// Dados de apoio para os testes
	validOrgID uuid.UUID
}

// SetupSuite: Roda uma vez antes de tudo. Sobe banco e configura a aplicação.
func (s *UserE2ESuite) SetupSuite() {
	cfg := config.Get()
	// Ajuste para Docker no Windows
	if cfg.DBHost == "localhost" {
		cfg.DBHost = "127.0.0.1"
	}

	// 1. Conecta no Banco (necessário para limpar tabelas)
	db, err := database.NewPostgres(cfg)
	s.Require().NoError(err)
	s.db = db

	// 2. Inicializa o Container (Injeção de Dependência Real)
	container, _, err := di.NewContainer(cfg)
	s.Require().NoError(err)
	s.container = container

	// 3. Inicializa o Router (A API completa)
	s.handler = routerLib.NewRouter(container)
}

// SetupTest: Roda antes de CADA teste. Limpa o banco e cria dados base.
func (s *UserE2ESuite) SetupTest() {
	// 1. Limpa tudo
	_, err := s.db.Exec("TRUNCATE organizations, users CASCADE")
	s.Require().NoError(err)

	// 2. CRIA UMA ORGANIZAÇÃO BASE (Necessária para criar usuário)
	// Usamos o Handler/UseCase de Organização que já está no container para facilitar
	// CNPJ Válido (Spotify Brasil): 15.502.834/0001-34
	orgInput := dto.CreateOrganizationRequest{
		Name:     "Empresa Teste User",
		Document: "15502834000134",
		Slug:     "empresa-teste-user",
		Sector:   "retail", // string direta ou entity.SectorRetail se acessível
		Plan:     "pro",
	}

	s.validOrgID = uuid.New()
	_, err = s.db.Exec(`
		INSERT INTO organizations (id, name, document, slug, plan, sector, settings, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, '{}', true)
	`, s.validOrgID, orgInput.Name, "15502834000134", orgInput.Slug, orgInput.Plan, orgInput.Sector)
	s.Require().NoError(err)
}

func (s *UserE2ESuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

// --- TESTES DE ROTA (E2E) ---

func (s *UserE2ESuite) TestRegisterAndLoginFlow() {
	// Variáveis do teste
	email := "usuario.teste@smartgondola.com"
	password := "SenhaForte123!"

	// ==========================================
	// 1. REGISTRAR USUÁRIO (POST /api/v1/auth/register)
	// ==========================================
	registerBody := userDTO.CreateUserRequest{
		OrganizationID: s.validOrgID, // <--- Agora temos um ID REAL!
		Name:           "João da Silva",
		Email:          email,
		Password:       password,
		Role:           entity.RoleManager,
	}
	body, _ := json.Marshal(registerBody)

	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handler.ServeHTTP(w, req)

	// Validações Registro
	s.Equal(http.StatusCreated, w.Code)

	var userResp userDTO.UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &userResp)
	s.NoError(err)
	s.Equal(email, userResp.Email)
	s.NotEmpty(userResp.ID)

	// ==========================================
	// 2. FAZER LOGIN (POST /api/v1/auth/login)
	// ==========================================
	loginBody := userDTO.LoginRequest{
		Email:    email,
		Password: password,
	}
	bodyLogin, _ := json.Marshal(loginBody)

	reqLogin, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyLogin))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()

	s.handler.ServeHTTP(wLogin, reqLogin)

	// Validações Login
	s.Equal(http.StatusOK, wLogin.Code)

	var loginResp map[string]interface{}
	err = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	s.NoError(err)
	s.NotEmpty(loginResp["access_token"], "Token não deve ser vazio")
}

func (s *UserE2ESuite) TestLogin_InvalidCredentials() {
	// Tenta logar sem criar usuário antes
	loginBody := userDTO.LoginRequest{
		Email:    "naoexiste@email.com",
		Password: "123",
	}
	body, _ := json.Marshal(loginBody)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handler.ServeHTTP(w, req)

	s.NotEqual(http.StatusOK, w.Code)
}

func TestUserE2ESuite(t *testing.T) {
	suite.Run(t, new(UserE2ESuite))
}
