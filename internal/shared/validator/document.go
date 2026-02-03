package validator

import (
	"regexp"
	"strconv"
)

// IsCNPJ verifica se um string é um CNPJ válido
func IsCNPJ(cnpj string) bool {
	// Remove caracteres não numéricos
	reg := regexp.MustCompile("[^0-9]")
	cnpj = reg.ReplaceAllString(cnpj, "")

	// CNPJ deve ter 14 dígitos
	if len(cnpj) != 14 {
		return false
	}

	// Elimina CNPJs inválidos conhecidos (todos dígitos iguais)
	if cnpj == "00000000000000" || cnpj == "11111111111111" ||
		cnpj == "22222222222222" || cnpj == "33333333333333" ||
		cnpj == "44444444444444" || cnpj == "55555555555555" ||
		cnpj == "66666666666666" || cnpj == "77777777777777" ||
		cnpj == "88888888888888" || cnpj == "99999999999999" {
		return false
	}

	// Validação do primeiro dígito verificador
	tamanho := 12
	numeros := cnpj[:tamanho]
	digitos := cnpj[tamanho:]
	soma := 0
	pos := 5

	for i := 0; i < tamanho; i++ {
		num, _ := strconv.Atoi(string(numeros[i]))
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

	if strconv.Itoa(resultado) != string(digitos[0]) {
		return false
	}

	// Validação do segundo dígito verificador
	tamanho = 13
	numeros = cnpj[:tamanho]
	soma = 0
	pos = 6

	for i := 0; i < tamanho; i++ {
		num, _ := strconv.Atoi(string(numeros[i]))
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

	if strconv.Itoa(resultado) != string(digitos[1]) {
		return false
	}

	return true
}
