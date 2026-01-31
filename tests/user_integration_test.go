package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/application/usecase"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/domain/entity"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/infrastructure/repository"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/interface/http/handler"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/database"
)

// setupTest sobe a infraestrutura necessária para o teste
func setupTest(t *testing.T) (*handler.UserHandler, *chi.Mux) {
	// 1. Carrega Config (assume que o .env está na raiz ou variáveis de ambiente setadas)
	cfg := config.Get()

	// 2. Conecta no Banco (Requer Docker rodando)
	db, err := database.NewPostgres(cfg)
	if err != nil {
		t.Fatalf("Erro ao conectar no banco de testes: %v", err)
	}

	// 3. Monta as dependências
	repo := repository.NewUserRepository(db)
	uc := usecase.NewUserUseCase(repo)
	h := handler.NewUserHandler(uc)

	// 4. Configura o Router
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	return h, r
}

func TestUserFlow_RegisterAndLogin(t *testing.T) {
	// Inicializa o ambiente
	_, router := setupTest(t)

	// Gera dados aleatórios para não dar conflito de Email Unique
	randomEmail := fmt.Sprintf("teste_%d@smartgondola.com", time.Now().UnixNano())
	randomPass := "senha123"

	// ==========================================
	// TESTE 1: REGISTRAR USUÁRIO (POST /auth/register)
	// ==========================================
	t.Run("Deve criar um usuário com sucesso", func(t *testing.T) {
		// Payload (Corpo da requisição)
		reqBody := dto.CreateUserRequest{
			OrganizationID: uuid.New(),
			Name:           "Usuario Teste Integração",
			Email:          randomEmail,
			Password:       randomPass,
			Role:           entity.RoleManager,
		}
		jsonBody, _ := json.Marshal(reqBody)

		// Cria a requisição HTTP
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Cria o Gravador de Resposta (ResponseRecorder)
		rr := httptest.NewRecorder()

		// Executa a requisição no router
		router.ServeHTTP(rr, req)

		// VALIDAÇÕES
		assert.Equal(t, http.StatusCreated, rr.Code) // Espera 201 Created

		var resp dto.UserResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, randomEmail, resp.Email)
	})

	// ==========================================
	// TESTE 2: FAZER LOGIN (POST /auth/login)
	// ==========================================
	t.Run("Deve fazer login e retornar token", func(t *testing.T) {
		// Payload
		loginBody := dto.LoginRequest{
			Email:    randomEmail,
			Password: randomPass,
		}
		jsonBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// VALIDAÇÕES
		assert.Equal(t, http.StatusOK, rr.Code) // Espera 200 OK

		var loginResp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &loginResp)
		assert.NoError(t, err)

		// Verifica se veio o token
		assert.NotEmpty(t, loginResp["access_token"])
		assert.Equal(t, "Bearer", loginResp["token_type"])
	})

	// ==========================================
	// TESTE 3: FALHA DE LOGIN (Senha Errada)
	// ==========================================
	t.Run("Deve negar login com senha errada", func(t *testing.T) {
		loginBody := dto.LoginRequest{
			Email:    randomEmail,
			Password: "senha_errada_123",
		}
		jsonBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code) // Espera 401 Unauthorized
	})
}
