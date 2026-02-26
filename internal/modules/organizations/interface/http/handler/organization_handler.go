package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/response"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/usecase"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/validator"
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
		response.Error(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if validationErrors := validator.ValidateStruct(req); len(validationErrors) > 0 {
		response.Error(w, http.StatusBadRequest, "Falha na validação dos dados", validationErrors...)
		return
	}

	// 2. Chama UseCase
	res, err := h.useCase.Create(r.Context(), req)
	if err != nil {
		// Tratamento de erros específicos
		if err.Error() == "este slug já está em uso por outra empresa" {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		if err.Error() == "CNPJ inválido" || err.Error() == "setor de atuação inválido" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 3. Responde 201 Created
	response.Created(w, res)
}

// GetByID trata GET /organizations/{id}
func (h *OrganizationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// 1. Pega ID da URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	// 2. Chama UseCase
	res, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Organização não encontrada")
		return
	}

	// 3. Responde 200 OK
	response.OK(w, res)
}

// Update trata PUT /organizations/{id}
func (h *OrganizationHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req dto.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if validationErrors := validator.ValidateStruct(req); len(validationErrors) > 0 {
		response.Error(w, http.StatusBadRequest, "Falha na validação dos dados", validationErrors...)
		return
	}

	res, err := h.useCase.Update(r.Context(), id, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(w, res)
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
