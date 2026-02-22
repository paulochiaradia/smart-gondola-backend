package pagination

import (
	"math"
	"net/http"
	"strconv"
)

// Params define os parâmetros de entrada da paginação
type Params struct {
	Page  int
	Limit int
}

// Meta define os metadados que serão devolvidos no JSON de resposta
type Meta struct {
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	Page       int   `json:"current_page"`
	Limit      int   `json:"limit"`
}

// NewParams extrai os parâmetros da URL com valores padrão (Page 1, Limit 10)
func NewParams(r *http.Request) Params {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10 // Padrão: 10 itens por página
	}

	// Trava de segurança para não pedirem 1 milhão de itens de uma vez
	if limit > 100 {
		limit = 100
	}

	return Params{
		Page:  page,
		Limit: limit,
	}
}

// Offset calcula quantos itens pular no banco de dados
func (p Params) Offset() int {
	return (p.Page - 1) * p.Limit
}

// NewMeta calcula o total de páginas baseado no total de itens e no limite
func NewMeta(totalItems int64, page, limit int) Meta {
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	return Meta{
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}
}
