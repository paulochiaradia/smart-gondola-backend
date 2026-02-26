package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// validate é a instância única (singleton) do validador em memória
var validate = validator.New()

// ValidateStruct recebe qualquer struct (DTO) e retorna uma lista de mensagens de erro
func ValidateStruct(s interface{}) []string {
	var errors []string

	err := validate.Struct(s)
	if err != nil {
		// Itera sobre todos os erros de validação encontrados na struct
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, formatError(err))
		}
	}

	return errors
}

// formatError traduz as tags do validador para mensagens legíveis
func formatError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("O campo '%s' é obrigatório", err.Field())
	case "email":
		return fmt.Sprintf("O campo '%s' deve ser um e-mail válido", err.Field())
	case "min":
		return fmt.Sprintf("O campo '%s' deve ter no mínimo %s caracteres", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("O campo '%s' deve ter no máximo %s caracteres", err.Field(), err.Param())
	case "oneof":
		return fmt.Sprintf("O campo '%s' deve ser um dos seguintes valores: %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("O campo '%s' é inválido", err.Field())
	}
}
