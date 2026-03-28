package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/paulochiaradia/smart-gondola-backend/internal/di"
	routerLib "github.com/paulochiaradia/smart-gondola-backend/internal/interface/http"
	"github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/response"
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
	email := "usuario.teste@smartgondola.com"
	password := "SenhaForte123!"

	// ==========================================
	// 1. REGISTRAR USUÁRIO
	// ==========================================
	registerBody := userDTO.CreateUserRequest{
		OrganizationID: s.validOrgID,
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

	// NOVO: Criamos uma struct anônima para ler o envelope "data"
	var registerResp struct {
		Data userDTO.UserResponse `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &registerResp)
	s.NoError(err)
	s.Equal(email, registerResp.Data.Email)
	s.NotEmpty(registerResp.Data.ID)

	// ==========================================
	// 2. FAZER LOGIN
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

	// NOVO: Lendo o envelope "data" do login
	var loginResp struct {
		Data map[string]interface{} `json:"data"`
	}
	err = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	s.NoError(err)
	s.NotEmpty(loginResp.Data["access_token"], "Token não deve ser vazio")
}
func (s *UserE2ESuite) TestLogin_InvalidCredentials() {
	loginBody := userDTO.LoginRequest{
		Email:    "naoexiste@email.com",
		Password: "123",
	}
	body, _ := json.Marshal(loginBody)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handler.ServeHTTP(w, req)

	// Agora esperamos 401
	s.Equal(http.StatusUnauthorized, w.Code)

	// NOVO: Verifica se o erro veio envelopado certinho
	var errorResp struct {
		Error response.ErrorPayload `json:"error"` // Importe o pacote response lá em cima se precisar
	}
	json.Unmarshal(w.Body.Bytes(), &errorResp)
	s.Equal(http.StatusUnauthorized, errorResp.Error.Code)
	s.NotEmpty(errorResp.Error.Message)
}

func (s *UserE2ESuite) TestLogin_LockoutMechanism() {
	// 1. Prepara os dados do alvo (você pode usar um usuário criado no SetupTest,
	// mas vamos assumir que existe um email válido e uma senha correta)
	email := "alvo_bruteforce@smartgondola.com"
	correctPassword := "SenhaSegura123!"

	// (Simulando o registro rápido desse usuário alvo para o teste)
	registerReq := userDTO.CreateUserRequest{
		OrganizationID: s.validOrgID, // Assumindo que você tem isso no seu Suite
		Name:           "Alvo Lockout",
		Email:          email,
		Password:       correctPassword,
		Role:           entity.RoleManager,
	}
	bodyReg, _ := json.Marshal(registerReq)
	reqReg, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(bodyReg))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	s.handler.ServeHTTP(wReg, reqReg)
	s.Require().Equal(http.StatusCreated, wReg.Code)

	// 2. O Atacante tenta errar a senha 5 vezes seguidas
	for i := 1; i <= 5; i++ {
		loginReq := userDTO.LoginRequest{
			Email:    email,
			Password: "SenhaIncorreta", // Errou!
		}
		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.handler.ServeHTTP(w, req)

		if i < 5 {
			s.Equal(http.StatusUnauthorized, w.Code)
		} else {
			s.Equal(http.StatusTooManyRequests, w.Code)
		}
	}

	// 3. A 6ª tentativa DEVE informar que a conta está bloqueada,
	// MESMO QUE ele finalmente acerte a senha!
	loginReq := userDTO.LoginRequest{
		Email:    email,
		Password: correctPassword, // Acertou, mas é tarde demais!
	}
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, req)

	// Conta bloqueada deve responder 429 mesmo com senha correta.
	s.Equal(http.StatusTooManyRequests, w.Code)

	var errorResp struct {
		Error response.ErrorPayload `json:"error"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal(http.StatusTooManyRequests, errorResp.Error.Code)
	s.Contains(errorResp.Error.Message, "conta temporariamente bloqueada")
}

func (s *UserE2ESuite) TestLogin_LockoutExpired_AllowsLogin() {
	email := "expira_lockout@smartgondola.com"
	correctPassword := "SenhaSegura123!"

	registerReq := userDTO.CreateUserRequest{
		OrganizationID: s.validOrgID,
		Name:           "Usuário Expira Lockout",
		Email:          email,
		Password:       correctPassword,
		Role:           entity.RoleManager,
	}
	bodyReg, _ := json.Marshal(registerReq)
	reqReg, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(bodyReg))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	s.handler.ServeHTTP(wReg, reqReg)
	s.Require().Equal(http.StatusCreated, wReg.Code)

	for i := 1; i <= 5; i++ {
		loginReq := userDTO.LoginRequest{
			Email:    email,
			Password: "SenhaIncorreta",
		}
		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.handler.ServeHTTP(w, req)

		if i < 5 {
			s.Equal(http.StatusUnauthorized, w.Code)
		} else {
			s.Equal(http.StatusTooManyRequests, w.Code)
		}
	}

	_, err := s.db.Exec(`
		UPDATE users
		SET locked_until = NOW() - INTERVAL '1 minute'
		WHERE email = $1
	`, email)
	s.Require().NoError(err)

	loginReq := userDTO.LoginRequest{
		Email:    email,
		Password: correctPassword,
	}
	bodyLogin, _ := json.Marshal(loginReq)
	reqLogin, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyLogin))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()

	s.handler.ServeHTTP(wLogin, reqLogin)
	s.Equal(http.StatusOK, wLogin.Code)

	var loginResp struct {
		Data map[string]interface{} `json:"data"`
	}
	err = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	s.NoError(err)
	s.NotEmpty(loginResp.Data["access_token"], "token deve existir após expiração do lockout")
}

func (s *UserE2ESuite) TestLogin_SuccessResetsSecurityFields() {
	email := "reset_security_fields@smartgondola.com"
	correctPassword := "SenhaSegura123!"

	registerReq := userDTO.CreateUserRequest{
		OrganizationID: s.validOrgID,
		Name:           "Usuário Reset Segurança",
		Email:          email,
		Password:       correctPassword,
		Role:           entity.RoleManager,
	}
	bodyReg, _ := json.Marshal(registerReq)
	reqReg, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(bodyReg))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	s.handler.ServeHTTP(wReg, reqReg)
	s.Require().Equal(http.StatusCreated, wReg.Code)

	for i := 0; i < 2; i++ {
		badLogin := userDTO.LoginRequest{Email: email, Password: "SenhaIncorreta"}
		bodyBad, _ := json.Marshal(badLogin)
		reqBad, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyBad))
		reqBad.Header.Set("Content-Type", "application/json")
		wBad := httptest.NewRecorder()
		s.handler.ServeHTTP(wBad, reqBad)
		s.Equal(http.StatusUnauthorized, wBad.Code)
	}

	goodLogin := userDTO.LoginRequest{Email: email, Password: correctPassword}
	bodyGood, _ := json.Marshal(goodLogin)
	reqGood, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyGood))
	reqGood.Header.Set("Content-Type", "application/json")
	wGood := httptest.NewRecorder()
	s.handler.ServeHTTP(wGood, reqGood)
	s.Require().Equal(http.StatusOK, wGood.Code)

	var failedAttempts int
	var lockedUntil sql.NullTime
	var lastLoginAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT failed_login_attempts, locked_until, last_login_at
		FROM users
		WHERE email = $1
	`, email).Scan(&failedAttempts, &lockedUntil, &lastLoginAt)
	s.Require().NoError(err)

	s.Equal(0, failedAttempts)
	s.False(lockedUntil.Valid)
	s.True(lastLoginAt.Valid)
}

func (s *UserE2ESuite) TestLogin_InactiveUser_ShouldReturnUnauthorized() {
	email := "inactive_user@smartgondola.com"
	password := "SenhaSegura123!"

	registerReq := userDTO.CreateUserRequest{
		OrganizationID: s.validOrgID,
		Name:           "Usuário Inativo",
		Email:          email,
		Password:       password,
		Role:           entity.RoleManager,
	}
	bodyReg, _ := json.Marshal(registerReq)
	reqReg, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(bodyReg))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	s.handler.ServeHTTP(wReg, reqReg)
	s.Require().Equal(http.StatusCreated, wReg.Code)

	_, err := s.db.Exec(`
		UPDATE users
		SET status = 'suspended'
		WHERE email = $1
	`, email)
	s.Require().NoError(err)

	loginReq := userDTO.LoginRequest{Email: email, Password: password}
	bodyLogin, _ := json.Marshal(loginReq)
	reqLogin, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyLogin))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()

	s.handler.ServeHTTP(wLogin, reqLogin)
	s.Equal(http.StatusUnauthorized, wLogin.Code)

	var errorResp struct {
		Error response.ErrorPayload `json:"error"`
	}
	err = json.Unmarshal(wLogin.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal(http.StatusUnauthorized, errorResp.Error.Code)
	s.NotEmpty(errorResp.Error.Message)
}

func (s *UserE2ESuite) TestLogin_LockoutConcurrentAttempts() {
	email := "concorrente_lockout@smartgondola.com"
	password := "SenhaSegura123!"

	registerReq := userDTO.CreateUserRequest{
		OrganizationID: s.validOrgID,
		Name:           "Usuário Concorrência Lockout",
		Email:          email,
		Password:       password,
		Role:           entity.RoleManager,
	}
	bodyReg, _ := json.Marshal(registerReq)
	reqReg, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(bodyReg))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	s.handler.ServeHTTP(wReg, reqReg)
	s.Require().Equal(http.StatusCreated, wReg.Code)

	const attempts = 10
	var wg sync.WaitGroup
	results := make(chan int, attempts)

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loginReq := userDTO.LoginRequest{Email: email, Password: "SenhaIncorreta"}
			body, _ := json.Marshal(loginReq)
			req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			s.handler.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	wg.Wait()
	close(results)

	var has429 bool
	for code := range results {
		s.NotEqual(http.StatusOK, code)
		if code == http.StatusTooManyRequests {
			has429 = true
		}
	}

	if !has429 {
		for i := 0; i < 10; i++ {
			loginReq := userDTO.LoginRequest{Email: email, Password: "SenhaIncorreta"}
			body, _ := json.Marshal(loginReq)
			req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			s.handler.ServeHTTP(w, req)
			if w.Code == http.StatusTooManyRequests {
				has429 = true
				break
			}
		}
	}
	s.True(has429, "após burst concorrente, o lockout deve ocorrer de forma eventual")

	var failedAttempts int
	var lockedUntil sql.NullTime
	err := s.db.QueryRow(`
		SELECT failed_login_attempts, locked_until
		FROM users
		WHERE email = $1
	`, email).Scan(&failedAttempts, &lockedUntil)
	s.Require().NoError(err)
	s.GreaterOrEqual(failedAttempts, 1)
	s.True(lockedUntil.Valid)
}

func (s *UserE2ESuite) TestLogin_AfterLockoutExpiresAndSuccess_FirstFailureStartsFromOne() {
	email := "apos_expirar_reinicia_contador@smartgondola.com"
	password := "SenhaSegura123!"

	registerReq := userDTO.CreateUserRequest{
		OrganizationID: s.validOrgID,
		Name:           "Usuário Reinício Contador",
		Email:          email,
		Password:       password,
		Role:           entity.RoleManager,
	}
	bodyReg, _ := json.Marshal(registerReq)
	reqReg, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(bodyReg))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	s.handler.ServeHTTP(wReg, reqReg)
	s.Require().Equal(http.StatusCreated, wReg.Code)

	for i := 1; i <= 5; i++ {
		badLogin := userDTO.LoginRequest{Email: email, Password: "SenhaIncorreta"}
		bodyBad, _ := json.Marshal(badLogin)
		reqBad, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyBad))
		reqBad.Header.Set("Content-Type", "application/json")
		wBad := httptest.NewRecorder()
		s.handler.ServeHTTP(wBad, reqBad)
	}

	_, err := s.db.Exec(`
		UPDATE users
		SET locked_until = NOW() - INTERVAL '1 minute'
		WHERE email = $1
	`, email)
	s.Require().NoError(err)

	goodLogin := userDTO.LoginRequest{Email: email, Password: password}
	bodyGood, _ := json.Marshal(goodLogin)
	reqGood, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyGood))
	reqGood.Header.Set("Content-Type", "application/json")
	wGood := httptest.NewRecorder()
	s.handler.ServeHTTP(wGood, reqGood)
	s.Require().Equal(http.StatusOK, wGood.Code)

	postSuccessFail := userDTO.LoginRequest{Email: email, Password: "SenhaErradaDeNovo"}
	bodyFail, _ := json.Marshal(postSuccessFail)
	reqFail, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(bodyFail))
	reqFail.Header.Set("Content-Type", "application/json")
	wFail := httptest.NewRecorder()
	s.handler.ServeHTTP(wFail, reqFail)
	s.Equal(http.StatusUnauthorized, wFail.Code)

	var failedAttempts int
	var lockedUntil sql.NullTime
	err = s.db.QueryRow(`
		SELECT failed_login_attempts, locked_until
		FROM users
		WHERE email = $1
	`, email).Scan(&failedAttempts, &lockedUntil)
	s.Require().NoError(err)
	s.Equal(1, failedAttempts)
	s.False(lockedUntil.Valid)
}

func TestUserE2ESuite(t *testing.T) {
	suite.Run(t, new(UserE2ESuite))
}
