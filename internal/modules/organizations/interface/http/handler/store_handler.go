package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/usecase"
)

type StoreHandler struct {
	useCase *usecase.StoreUseCase
}

func NewStoreHandler(uc *usecase.StoreUseCase) *StoreHandler {
	return &StoreHandler{useCase: uc}
}

// Create POST /stores
func (h *StoreHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	res, err := h.useCase.Create(r.Context(), req)
	if err != nil {
		// Mapeando erros de negócio para Status Code
		if err.Error() == "já existe uma loja com este código nesta organização" {
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

// ListByOrg GET /organizations/{orgId}/stores
func (h *StoreHandler) ListByOrg(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "ID da organização inválido", http.StatusBadRequest)
		return
	}

	res, err := h.useCase.ListByOrganization(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// RegisterRoutes registra as rotas (será chamado pelo Router principal)
func (h *StoreHandler) RegisterRoutes(router chi.Router) {
	// Rotas diretas de Loja
	router.Post("/stores", h.Create)

	// Rotas aninhadas (sub-recurso de organização)
	// GET /organizations/{orgId}/stores
	router.Get("/organizations/{orgId}/stores", h.ListByOrg)
}
