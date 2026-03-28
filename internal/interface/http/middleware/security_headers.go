package middleware

import "net/http"

// SecurityHeaders adiciona cabeçalhos HTTP recomendados pela OWASP
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Impede que o navegador tente adivinhar o MIME type (sniffing)
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Impede que a API seja colocada dentro de um <iframe / <frame> (Clickjacking)
		w.Header().Set("X-Frame-Options", "DENY")

		// Ativa a proteção XSS nativa dos navegadores mais antigos
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Força conexões seguras (HSTS) - Máximo de 1 ano (31536000 segundos)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}
