package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/middleware"
	"github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/response"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/users/application/usecase"
)

type UserHandler struct {
	useCase *usecase.UserUseCase
}

// NewUserHandler cria o controller
func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{useCase: uc}
}

// Register trata a rota POST /auth/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest

	// 1. Decodifica o JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	// 2. Chama a regra de negócio
	res, err := h.useCase.Register(r.Context(), req)
	if err != nil {
		// Tratamento de erro simplificado
		if err.Error() == "email já cadastrado" {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Responde Sucesso (201 Created)
	response.Created(w, res)
}

// Login trata a rota POST /auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	res, err := h.useCase.Login(r.Context(), req)
	if err != nil {
		// Por segurança, sempre retorna 401 genérico
		response.Error(w, http.StatusUnauthorized, "email ou senha inválidos")
		return
	}

	response.OK(w, res)
}

// Me retorna os dados do usuário logado (extraídos do Token)
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Recupera o ID que o Middleware injetou no contexto
	userID := middleware.GetUserID(r.Context())
	orgID := middleware.GetOrgID(r.Context())
	role := middleware.GetRole(r.Context())

	// Retorna os dados do usuário autenticado
	data := map[string]interface{}{
		"id":      userID,
		"org_id":  orgID,
		"role":    role,
		"message": "Token válido! Você está autenticado.",
	}

	response.OK(w, data)
}

// RegisterRoutes agrupa as rotas do módulo
func (h *UserHandler) RegisterRoutes(router chi.Router) {
	// Rotas Públicas
	router.Post("/auth/register", h.Register)
	router.Post("/auth/login", h.Login)
}
