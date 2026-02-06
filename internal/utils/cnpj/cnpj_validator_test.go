package cnpj

import (
	"testing"
)

func TestValidateCNPJ(t *testing.T) {
	tests := []struct {
		name    string
		cnpj    string
		wantErr bool
	}{
		{
			name:    "CNPJ válido formatado",
			cnpj:    "11.222.333/0001-81",
			wantErr: false,
		},
		{
			name:    "CNPJ válido sem formatação",
			cnpj:    "11222333000181",
			wantErr: false,
		},
		{
			name:    "CNPJ inválido - menos de 14 dígitos",
			cnpj:    "1122233300018",
			wantErr: true,
		},
		{
			name:    "CNPJ inválido - mais de 14 dígitos",
			cnpj:    "112223330001811",
			wantErr: true,
		},
		{
			name:    "CNPJ inválido - todos dígitos iguais",
			cnpj:    "11.111.111/1111-11",
			wantErr: true,
		},
		{
			name:    "CNPJ inválido - dígito verificador errado",
			cnpj:    "11.222.333/0001-80",
			wantErr: true,
		},
		{
			name:    "CNPJ com caracteres inválidos",
			cnpj:    "11.222.333/0001-8A",
			wantErr: true,
		},
		{
			name:    "CNPJ vazio",
			cnpj:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCNPJ(tt.cnpj)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCNPJ() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatCNPJ(t *testing.T) {
	tests := []struct {
		name string
		cnpj string
		want string
	}{
		{
			name: "CNPJ sem formatação",
			cnpj: "11222333000181",
			want: "11.222.333/0001-81",
		},
		{
			name: "CNPJ já formatado",
			cnpj: "11.222.333/0001-81",
			want: "11.222.333/0001-81",
		},
		{
			name: "CNPJ parcialmente formatado",
			cnpj: "11222333/0001-81",
			want: "11.222.333/0001-81",
		},
		{
			name: "CNPJ inválido (mantém original)",
			cnpj: "123",
			want: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatCNPJ(tt.cnpj); got != tt.want {
				t.Errorf("FormatCNPJ() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAllSameDigit(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "Todos iguais",
			s:    "11111111111111",
			want: true,
		},
		{
			name: "Diferentes",
			s:    "11222333000181",
			want: false,
		},
		{
			name: "String vazia",
			s:    "",
			want: true,
		},
		{
			name: "Um caractere",
			s:    "1",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllSameDigit(tt.s); got != tt.want {
				t.Errorf("isAllSameDigit() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark para validação local de CNPJ
func BenchmarkValidateCNPJ(b *testing.B) {
	cnpj := "11.222.333/0001-81"
	for i := 0; i < b.N; i++ {
		_ = ValidateCNPJ(cnpj)
	}
}

// Benchmark para formatação de CNPJ
func BenchmarkFormatCNPJ(b *testing.B) {
	cnpj := "11222333000181"
	for i := 0; i < b.N; i++ {
		_ = FormatCNPJ(cnpj)
	}
}
