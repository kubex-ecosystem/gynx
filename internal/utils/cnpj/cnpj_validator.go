// Package cnpj fornece funcionalidades para validação e formatação de CNPJs.
package cnpj

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	gl "github.com/kubex-ecosystem/logz"
)

// ValidateCNPJ valida o formato e dígitos verificadores do CNPJ
func ValidateCNPJ(cnpj string) error {
	// Remove caracteres não numéricos
	cnpj = strings.ReplaceAll(cnpj, ".", "")
	cnpj = strings.ReplaceAll(cnpj, "/", "")
	cnpj = strings.ReplaceAll(cnpj, "-", "")
	cnpj = strings.TrimSpace(cnpj)

	// Verifica se tem 14 dígitos
	if len(cnpj) != 14 {
		return fmt.Errorf("CNPJ deve conter 14 dígitos")
	}

	// Verifica se todos são dígitos
	if matched, _ := regexp.MatchString(`^\d{14}$`, cnpj); !matched {
		return fmt.Errorf("CNPJ deve conter apenas números")
	}

	// Verifica CNPJs inválidos conhecidos (todos dígitos iguais)
	if isAllSameDigit(cnpj) {
		return fmt.Errorf("CNPJ inválido")
	}

	// Valida primeiro dígito verificador
	if !validateDigit(cnpj, 12) {
		return fmt.Errorf("CNPJ inválido: primeiro dígito verificador incorreto")
	}

	// Valida segundo dígito verificador
	if !validateDigit(cnpj, 13) {
		return fmt.Errorf("CNPJ inválido: segundo dígito verificador incorreto")
	}

	return nil
}

// FormatCNPJ formata o CNPJ no padrão XX.XXX.XXX/XXXX-XX
func FormatCNPJ(cnpj string) string {
	// Remove formatação existente
	cnpj = strings.ReplaceAll(cnpj, ".", "")
	cnpj = strings.ReplaceAll(cnpj, "/", "")
	cnpj = strings.ReplaceAll(cnpj, "-", "")
	cnpj = strings.TrimSpace(cnpj)

	if len(cnpj) != 14 {
		return cnpj
	}

	// Formata: XX.XXX.XXX/XXXX-XX
	return fmt.Sprintf("%s.%s.%s/%s-%s",
		cnpj[0:2],
		cnpj[2:5],
		cnpj[5:8],
		cnpj[8:12],
		cnpj[12:14],
	)
}

// isAllSameDigit verifica se todos os dígitos são iguais
func isAllSameDigit(s string) bool {
	if len(s) <= 1 {
		return true
	}
	first := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != first {
			return false
		}
	}
	return true
}

// validateDigit valida o dígito verificador na posição especificada
func validateDigit(cnpj string, position int) bool {
	var sum int
	var weight int

	if position == 12 {
		weight = 5
	} else {
		weight = 6
	}

	for i := 0; i < position; i++ {
		digit, _ := strconv.Atoi(string(cnpj[i]))
		sum += digit * weight
		weight--
		if weight < 2 {
			weight = 9
		}
	}

	remainder := sum % 11
	var expectedDigit int
	if remainder < 2 {
		expectedDigit = 0
	} else {
		expectedDigit = 11 - remainder
	}

	actualDigit, _ := strconv.Atoi(string(cnpj[position]))
	return actualDigit == expectedDigit
}

// ReceitaWSResponse representa a resposta da API ReceitaWS
type ReceitaWSResponse struct {
	CNPJ               string `json:"cnpj"`
	RazaoSocial        string `json:"nome"`
	NomeFantasia       string `json:"fantasia"`
	Abertura           string `json:"abertura"`
	Situacao           string `json:"situacao"`
	Tipo               string `json:"tipo"`
	Porte              string `json:"porte"`
	NaturezaJuridica   string `json:"natureza_juridica"`
	AtividadePrincipal []struct {
		Code string `json:"code"`
		Text string `json:"text"`
	} `json:"atividade_principal"`
	Logradouro     string `json:"logradouro"`
	Numero         string `json:"numero"`
	Complemento    string `json:"complemento"`
	Municipio      string `json:"municipio"`
	Bairro         string `json:"bairro"`
	UF             string `json:"uf"`
	CEP            string `json:"cep"`
	Email          string `json:"email"`
	Telefone       string `json:"telefone"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	DataSituacao   string `json:"data_situacao"`
	MotivoSituacao string `json:"motivo_situacao"`
	CapitalSocial  string `json:"capital_social"`
}

// ValidateWithReceitaWS valida CNPJ consultando a API da ReceitaWS
// Rate limit: 3 requisições por minuto (gratuito)
func ValidateWithReceitaWS(cnpj string) (*ReceitaWSResponse, error) {
	// Primeiro valida o formato localmente
	if err := ValidateCNPJ(cnpj); err != nil {
		return nil, err
	}

	// Remove formatação para a consulta
	cnpjClean := strings.ReplaceAll(cnpj, ".", "")
	cnpjClean = strings.ReplaceAll(cnpjClean, "/", "")
	cnpjClean = strings.ReplaceAll(cnpjClean, "-", "")

	// Consulta a API ReceitaWS
	url := fmt.Sprintf("https://www.receitaws.com.br/v1/cnpj/%s", cnpjClean)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, gl.Errorf("erro ao criar requisição: %v", err)
	}

	// Headers recomendados
	req.Header.Set("User-Agent", "Kubex-PRM/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, gl.Errorf("erro ao consultar ReceitaWS: %v", err)
	}
	defer resp.Body.Close()

	// Rate limit atingido (429)
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, gl.Errorf("limite de requisições atingido (3/min). Tente novamente em alguns segundos")
	}

	// Erro na API
	if resp.StatusCode != http.StatusOK {
		return nil, gl.Errorf("erro na API ReceitaWS: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, gl.Errorf("erro ao ler resposta: %v", err)
	}

	var result ReceitaWSResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, gl.Errorf("erro ao parsear JSON: %v", err)
	}

	// Verifica se o CNPJ existe
	if result.Status == "ERROR" {
		return nil, gl.Errorf("CNPJ não encontrado na Receita Federal: %s", result.Message)
	}

	// Verifica situação cadastral
	if result.Situacao != "ATIVA" {
		return nil, gl.Errorf("CNPJ com situação cadastral '%s': %s",
			result.Situacao, result.MotivoSituacao)
	}

	return &result, nil
}

// CNPJInfo contém informações resumidas do CNPJ
type CNPJInfo struct {
	CNPJ         string
	RazaoSocial  string
	NomeFantasia string
	Situacao     string
	Cidade       string
	UF           string
	IsValid      bool
	Message      string
}

// GetCNPJInfo retorna informações resumidas do CNPJ (wrapper simplificado)
func GetCNPJInfo(cnpj string) (*CNPJInfo, error) {
	result, err := ValidateWithReceitaWS(cnpj)
	if err != nil {
		return &CNPJInfo{
			CNPJ:    cnpj,
			IsValid: false,
			Message: err.Error(),
		}, err
	}

	return &CNPJInfo{
		CNPJ:         FormatCNPJ(result.CNPJ),
		RazaoSocial:  result.RazaoSocial,
		NomeFantasia: result.NomeFantasia,
		Situacao:     result.Situacao,
		Cidade:       result.Municipio,
		UF:           result.UF,
		IsValid:      true,
		Message:      "CNPJ válido e ativo",
	}, nil
}
