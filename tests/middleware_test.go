package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/paulochiaradia/smart-gondola-backend/internal/di"
	routerLib "github.com/paulochiaradia/smart-gondola-backend/internal/interface/http"
	"github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/middleware"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/config"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/database"
)

type MiddlewareSuite struct {
	suite.Suite
	handler http.Handler
}

func (s *MiddlewareSuite) SetupSuite() {
	cfg := config.Get()
	if cfg.DBHost == "localhost" {
		cfg.DBHost = "127.0.0.1"
	}

	db, _ := database.NewPostgres(cfg) // Ignora erro, pois o middleware roda antes do banco
	container, _, _ := di.NewContainer(cfg)

	if container != nil {
		container.DB = db
		s.handler = routerLib.NewRouter(container)
	}
}

func (s *MiddlewareSuite) TestProtectedRoutes_ShouldBlockAnonymous() {
	// Tenta acessar rota de criar loja sem token
	req, _ := http.NewRequest("POST", "/api/v1/stores", nil)
	w := httptest.NewRecorder()

	s.handler.ServeHTTP(w, req)

	// DEVE retornar 401 Unauthorized
	assert.Equal(s.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(s.T(), w.Body.String(), "Authorization header is required")
}

func (s *MiddlewareSuite) TestPublicRoutes_ShouldAllowAnonymous() {
	// Tenta acessar Health Check
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	s.handler.ServeHTTP(w, req)

	// DEVE retornar 200 OK
	assert.Equal(s.T(), http.StatusOK, w.Code)
}

// --- (Adicione os imports necessários lá em cima se faltar algum: "context", "github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/middleware") ---

func (s *MiddlewareSuite) TestRBAC_ShouldBlockOperatorFromCreatingStore() {
	// 1. Montamos uma requisição para criar loja
	req, _ := http.NewRequest("POST", "/api/v1/stores", nil)

	// 2. Simulamos um usuário com a role "operator" já logado (Mock do contexto)
	ctx := context.WithValue(req.Context(), middleware.RoleContextKey, "operator")
	req = req.WithContext(ctx)

	// Precisamos também simular um header de Authorization para passar pelo AuthMiddleware
	// (Num teste de integração completo, geraríamos um token real,
	// mas para não complicar, testamos o comportamento do fluxo bloqueando a requisição).
	// *Para simplificar*, vamos testar direto a resposta da requisição. Se seu AuthMiddleware
	// bloqueia antes por falta de token válido, o teste real seria gerar um token JWT com role 'operator'.

	// Para manter este teste focado apenas no RBAC sem precisar gerar JWT real no teste:
	// Vamos criar um Handler fake encapsulado pelo RequireRole
	rbacMiddleware := middleware.RequireRole("admin", "tenant")
	fakeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Se passar pelo RBAC, dá 200
	})

	protectedHandler := rbacMiddleware(fakeHandler)
	w := httptest.NewRecorder()

	// 3. Executamos o request com o contexto de "operator"
	protectedHandler.ServeHTTP(w, req)

	// 4. DEVE retornar 403 Forbidden
	s.Equal(http.StatusForbidden, w.Code)
	s.Contains(w.Body.String(), "Acesso negado")
}

func (s *MiddlewareSuite) TestRBAC_ShouldAllowTenantToCreateStore() {
	req, _ := http.NewRequest("POST", "/api/v1/stores", nil)

	// Simulamos um usuário com a role "tenant"
	ctx := context.WithValue(req.Context(), middleware.RoleContextKey, "tenant")
	req = req.WithContext(ctx)

	rbacMiddleware := middleware.RequireRole("admin", "tenant")
	fakeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // Se passar pelo RBAC, dá 200
	})

	protectedHandler := rbacMiddleware(fakeHandler)
	w := httptest.NewRecorder()

	protectedHandler.ServeHTTP(w, req)

	// 4. DEVE retornar 200 OK (Passou pelo bloqueio!)
	s.Equal(http.StatusOK, w.Code)
}

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareSuite))
}
