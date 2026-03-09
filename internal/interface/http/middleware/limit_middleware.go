package middleware

import (
	"net/http"
)

// LimitPayloadSize restringe o tamanho máximo do corpo da requisição
func LimitPayloadSize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limita o tamanho do body. Se passar disso, o r.Body retornará um erro
			// quando o json.NewDecoder tentar ler, acionando nosso erro 400 padronizado.
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

			next.ServeHTTP(w, r)
		})
	}
}
