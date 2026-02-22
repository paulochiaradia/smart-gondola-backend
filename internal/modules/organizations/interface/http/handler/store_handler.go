package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paulochiaradia/smart-gondola-backend/internal/interface/http/response"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/dto"
	"github.com/paulochiaradia/smart-gondola-backend/internal/modules/organizations/application/usecase"
	"github.com/paulochiaradia/smart-gondola-backend/internal/shared/pagination"
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
		response.Error(w, http.StatusBadRequest, "Formato JSON inválido")
		return
	}

	res, err := h.useCase.Create(r.Context(), req)
	if err != nil {
		if err.Error() == "já existe uma loja com este código nesta organização" {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Erro interno ao criar loja", err.Error())
		return
	}

	response.Created(w, res)
}

// ListByOrg GET /organizations/{orgId}/stores
func (h *StoreHandler) ListByOrg(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "ID da organização inválido")
		return
	}

	// 1. Extrai a página e o limite da URL (ex: ?page=1&limit=10)
	pageParams := pagination.NewParams(r)

	// 2. Chama o UseCase passando os parâmetros e recebendo as 3 variáveis!
	res, totalItems, err := h.useCase.ListByOrganization(r.Context(), orgID, pageParams)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Erro ao buscar lojas", err.Error())
		return
	}

	// 3. Monta os Metadados (total de páginas, itens, etc)
	meta := pagination.NewMeta(totalItems, pageParams.Page, pageParams.Limit)

	// 4. Devolve JSON padronizado com Data e Meta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response.SuccessPayload{
		Data: res,
		Meta: meta,
	})
}

func (h *StoreHandler) RegisterRoutes(router chi.Router) {
	router.Post("/stores", h.Create)
	router.Get("/organizations/{orgId}/stores", h.ListByOrg)
}
