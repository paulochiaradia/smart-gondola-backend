package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/usecase"
)

type OrganizationHandler struct {
	useCase *usecase.OrganizationUseCase
}

// NewOrganizationHandler cria o controller
func NewOrganizationHandler(uc *usecase.OrganizationUseCase) *OrganizationHandler {
	return &OrganizationHandler{useCase: uc}
}

// Create trata POST /organizations
func (h *OrganizationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrganizationRequest

	// 1. Decodifica JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 2. Chama UseCase
	res, err := h.useCase.Create(r.Context(), req)
	if err != nil {
		// Tratamento de erros específicos
		if err.Error() == "este slug já está em uso por outra empresa" {
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
			return
		}
		if err.Error() == "CNPJ inválido" || err.Error() == "setor de atuação inválido" {
			http.Error(w, err.Error(), http.StatusBadRequest) // 400 Bad Request
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Responde 201 Created
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// GetByID trata GET /organizations/{id}
func (h *OrganizationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// 1. Pega ID da URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// 2. Chama UseCase
	res, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Organização não encontrada", http.StatusNotFound)
		return
	}

	// 3. Responde 200 OK
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Update trata PUT /organizations/{id}
func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req dto.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	res, err := h.useCase.Update(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// RegisterRoutes registra as rotas no router principal
func (h *OrganizationHandler) RegisterRoutes(router chi.Router) {
	// Agrupamento /organizations
	router.Route("/organizations", func(r chi.Router) {
		r.Post("/", h.Create)     // Criar empresa
		r.Get("/{id}", h.GetByID) // Buscar empresa
		r.Put("/{id}", h.Update)  // Atualizar dados
	})
}
