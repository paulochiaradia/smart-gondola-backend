package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/middleware"
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
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 2. Chama a regra de negócio
	res, err := h.useCase.Register(r.Context(), req)
	if err != nil {
		// Tratamento de erro simplificado
		if err.Error() == "email já cadastrado" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Responde Sucesso (201 Created)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// Login trata a rota POST /auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	res, err := h.useCase.Login(r.Context(), req)
	if err != nil {
		// Por segurança, sempre retorna 401 genérico
		http.Error(w, "email ou senha inválidos", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// Me retorna os dados do usuário logado (extraídos do Token)
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Recupera o ID que o Middleware injetou no contexto
	userID := middleware.GetUserID(r.Context())
	orgID := middleware.GetOrgID(r.Context())
	role := middleware.GetRole(r.Context())

	// Retorna um JSON simples para validar
	response := map[string]interface{}{
		"id":      userID,
		"org_id":  orgID,
		"role":    role,
		"message": "Token válido! Você está autenticado.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes agrupa as rotas do módulo
func (h *UserHandler) RegisterRoutes(router chi.Router) {
	// Rotas Públicas
	router.Post("/auth/register", h.Register)
	router.Post("/auth/login", h.Login)
}
