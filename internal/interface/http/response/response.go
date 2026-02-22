package response

import (
	"encoding/json"
	"net/http"
)

// ErrorPayload define a estrutura padrão para qualquer erro na API
type ErrorPayload struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"` // Opcional: lista de erros (ex: validação de campos)
}

// SuccessPayload define a estrutura padrão para sucesso (prepara terreno para paginação)
type SuccessPayload struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta,omitempty"` // Metadados (ex: página atual, total de itens)
}

// Error envia uma resposta de erro padronizada em JSON
func Error(w http.ResponseWriter, statusCode int, message string, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := map[string]interface{}{
		"error": ErrorPayload{
			Code:    statusCode,
			Message: message,
			Details: details,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// JSON envia uma resposta de sucesso genérica com envelope "data"
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(SuccessPayload{Data: data})
	}
}

// OK é um atalho para 200 OK
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created é um atalho para 201 Created
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent envia um 204 Sem Conteúdo
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
