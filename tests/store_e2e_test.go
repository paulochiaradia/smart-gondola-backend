package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/paulochiaradia/smart-gondola-backend/internal/di"
	routerLib "github.com/paulochiaradia/smart-gondola-backend/internal/interface/http"
	"github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/response"
	orgDTO "github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	userDTO "github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/database"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/pagination"
)

type StoreE2ESuite struct {
	suite.Suite
	db         *sql.DB
	handler    http.Handler
	validOrgID uuid.UUID
	validToken string // Guardaremos o JWT real aqui
}

func (s *StoreE2ESuite) SetupSuite() {
	cfg := config.Get()
	if cfg.DBHost == "localhost" {
		cfg.DBHost = "127.0.0.1"
	}

	db, err := database.NewPostgres(cfg)
	s.Require().NoError(err)
	s.db = db

	container, _, err := di.NewContainer(cfg)
	s.Require().NoError(err)
	s.handler = routerLib.NewRouter(container)
}

// SetupTest roda antes de CADA teste. Limpa o banco, cria a Org, Registra Usuário e faz Login.
func (s *StoreE2ESuite) SetupTest() {
	// 1. Limpa as tabelas
	_, err := s.db.Exec("TRUNCATE organizations, users, stores CASCADE")
	s.Require().NoError(err)

	// 2. Cria a Organização base via SQL
	s.validOrgID = uuid.New()
	_, err = s.db.Exec(`
		INSERT INTO organizations (id, name, document, slug, plan, sector, settings, is_active)
		VALUES ($1, 'Org Teste', '11111111111111', 'org-teste', 'pro', 'retail', '{}', true)
	`, s.validOrgID)
	s.Require().NoError(err)

	// 3. Registra um Usuário "tenant" (Dono) via HTTP
	email := "dono@smartgondola.com"
	password := "SenhaForte123!"

	registerBody := userDTO.CreateUserRequest{
		OrganizationID: s.validOrgID,
		Name:           "Dono da Empresa",
		Email:          email,
		Password:       password,
		Role:           "tenant", // Role que tem permissão no RBAC
	}
	body, _ := json.Marshal(registerBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	// 4. Faz Login para pegar o Token JWT Real
	loginBody := userDTO.LoginRequest{
		Email:    email,
		Password: password,
	}
	bodyLogin, _ := json.Marshal(loginBody)
	reqLogin, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyLogin))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	s.handler.ServeHTTP(wLogin, reqLogin)
	s.Require().Equal(http.StatusOK, wLogin.Code)

	var loginResp struct {
		Data map[string]interface{} `json:"data"`
	}
	json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	s.validToken = loginResp.Data["access_token"].(string)
}

func (s *StoreE2ESuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

// --- TESTES DE ENDPOINT ---

func (s *StoreE2ESuite) TestStoreEndpoints_CreateAndListWithPagination() {
	// =========================================================
	// 1. CRIAR LOJA 1 (POST /stores) usando o Token
	// =========================================================
	store1 := orgDTO.CreateStoreRequest{
		OrganizationID: s.validOrgID,
		Name:           "Loja Matriz",
		Code:           "MATRIZ-01",
		Timezone:       "America/Sao_Paulo",
	}
	body1, _ := json.Marshal(store1)
	req1, _ := http.NewRequest("POST", "/api/v1/stores", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+s.validToken) // <-- Injetando o JWT

	w1 := httptest.NewRecorder()
	s.handler.ServeHTTP(w1, req1)

	// Deve retornar 201 Created com o formato de resposta padrão
	s.Equal(http.StatusCreated, w1.Code)

	var createResp struct {
		Data orgDTO.StoreResponse `json:"data"`
	}
	err := json.Unmarshal(w1.Body.Bytes(), &createResp)
	s.NoError(err)
	s.Equal("Loja Matriz", createResp.Data.Name)

	// =========================================================
	// 2. CRIAR LOJA 2 (Para testarmos a paginação depois)
	// =========================================================
	store2 := orgDTO.CreateStoreRequest{
		OrganizationID: s.validOrgID,
		Name:           "Loja Filial",
		Code:           "FILIAL-01",
		Timezone:       "America/Sao_Paulo",
	}
	body2, _ := json.Marshal(store2)
	req2, _ := http.NewRequest("POST", "/api/v1/stores", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+s.validToken)
	w2 := httptest.NewRecorder()
	s.handler.ServeHTTP(w2, req2)
	s.Equal(http.StatusCreated, w2.Code)

	// =========================================================
	// 3. LISTAR LOJAS COM PAGINAÇÃO (GET /stores?page=1&limit=1)
	// =========================================================
	url := fmt.Sprintf("/api/v1/organizations/%s/stores?page=1&limit=1", s.validOrgID)
	req3, _ := http.NewRequest("GET", url, nil)
	req3.Header.Set("Authorization", "Bearer "+s.validToken) // Sem token tomaria 401

	w3 := httptest.NewRecorder()
	s.handler.ServeHTTP(w3, req3)

	s.Equal(http.StatusOK, w3.Code)

	// Lendo a resposta que agora tem Data e Meta (Paginação)
	var listResp struct {
		Data []orgDTO.StoreResponse `json:"data"`
		Meta pagination.Meta        `json:"meta"`
	}
	err = json.Unmarshal(w3.Body.Bytes(), &listResp)
	s.NoError(err)

	// Validações da Paginação
	s.Len(listResp.Data, 1, "Deve retornar apenas 1 loja por causa do limit=1")
	s.Equal(int64(2), listResp.Meta.TotalItems, "O banco tem 2 lojas cadastradas no total")
	s.Equal(2, listResp.Meta.TotalPages, "Se o limite é 1 e tem 2 itens, devem ser 2 páginas")
}

func (s *StoreE2ESuite) TestStoreEndpoints_CreateWithValidationFailure() {
	// Manda requisição sem nome e sem código (Para acionar o Validator)
	invalidStore := orgDTO.CreateStoreRequest{
		OrganizationID: s.validOrgID,
		// Faltando campos obrigatórios
	}
	body, _ := json.Marshal(invalidStore)
	req, _ := http.NewRequest("POST", "/api/v1/stores", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.validToken)

	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, req)

	// Validador deve barrar na porta com HTTP 400
	s.Equal(http.StatusBadRequest, w.Code)

	var errResp struct {
		Error response.ErrorPayload `json:"error"`
	}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	s.Contains(errResp.Error.Message, "Falha na validação")
	s.True(len(errResp.Error.Details) > 0, "Deve listar os campos faltantes")
}

func TestStoreE2ESuite(t *testing.T) {
	suite.Run(t, new(StoreE2ESuite))
}
