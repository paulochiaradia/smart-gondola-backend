package validator

import (
	"regexp"
	"strconv"
)

// IsCNPJ verifica se um documento é válido (formato e dígitos verificadores)
func IsCNPJ(cnpj string) bool {
	// 1. Remove caracteres não numéricos
	reg, _ := regexp.Compile("[^0-9]")
	cnpj = reg.ReplaceAllString(cnpj, "")

	// 2. Verifica tamanho (CNPJ tem 14 dígitos)
	if len(cnpj) != 14 {
		return false
	}

	// 3. Elimina CNPJs inválidos conhecidos (00000000000000, etc)
	if isRepeated(cnpj) {
		return false
	}

	// 4. Valida os dígitos verificadores
	// Primeiro dígito
	tamanho := len(cnpj) - 2
	numeros := cnpj[0:tamanho]
	digitos := cnpj[tamanho:]
	soma := 0
	pos := tamanho - 7
	for i := tamanho; i >= 1; i-- {
		num, _ := strconv.Atoi(string(numeros[tamanho-i]))
		soma += num * pos
		pos--
		if pos < 2 {
			pos = 9
		}
	}
	resultado := soma % 11
	if resultado < 2 {
		resultado = 0
	} else {
		resultado = 11 - resultado
	}
	digito1, _ := strconv.Atoi(string(digitos[0]))
	if resultado != digito1 {
		return false
	}

	// Segundo dígito
	tamanho = tamanho + 1
	numeros = cnpj[0:tamanho]
	soma = 0
	pos = tamanho - 7
	for i := tamanho; i >= 1; i-- {
		num, _ := strconv.Atoi(string(numeros[tamanho-i]))
		soma += num * pos
		pos--
		if pos < 2 {
			pos = 9
		}
	}
	resultado = soma % 11
	if resultado < 2 {
		resultado = 0
	} else {
		resultado = 11 - resultado
	}
	digito2, _ := strconv.Atoi(string(digitos[1]))
	if resultado != digito2 {
		return false
	}

	return true
}

// Auxiliar para detectar "11111111111111"
func isRepeated(doc string) bool {
	first := doc[0]
	for i := 1; i < len(doc); i++ {
		if doc[i] != first {
			return false
		}
	}
	return true
}
