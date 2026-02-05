package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/paulochiaradia/smart-gondola-backend/internal/di"
	routerLib "github.com/paulochiaradia/smart-gondola-backend/internal/interface/http"
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

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareSuite))
}
