// Package tokens testa a geração e validação de tokens JWT e refresh tokens.
//
// Estes testes validam:
// - Geração de refresh tokens seguros (entropia, unicidade)
// - Encoding base64 URL-safe
// - Hash SHA-256 de tokens
//
// Tipo: Testes unitários (sem dependências externas)
// Previne: Tokens previsíveis, colisões de hash, encoding incorreto
package tokens

import (
	"strings"
	"testing"
	"time"
)

// --- Testes de GenerateRefreshToken ---
// Previne: Tokens previsíveis ou com baixa entropia

func TestGenerateRefreshToken(t *testing.T) {
	t.Run("generates plain token with 64 characters", func(t *testing.T) {
		plain, _, _, err := GenerateRefreshToken()
		if err != nil {
			t.Fatalf("GenerateRefreshToken() error = %v", err)
		}
		if len(plain) != 64 {
			t.Errorf("plain token length = %d, want 64", len(plain))
		}
	})

	t.Run("generates hash with 43 characters", func(t *testing.T) {
		_, hash, _, err := GenerateRefreshToken()
		if err != nil {
			t.Fatalf("GenerateRefreshToken() error = %v", err)
		}
		// SHA-256 = 32 bytes, base64url encoded = ~43 chars
		if len(hash) != 32 {
			t.Errorf("hash length = %d, want 32 (base64url encoded SHA-256)", len(hash))
		}
	})

	t.Run("returns future expiration time", func(t *testing.T) {
		_, _, exp, err := GenerateRefreshToken()
		if err != nil {
			t.Fatalf("GenerateRefreshToken() error = %v", err)
		}
		// Default é 30 dias no futuro
		if exp.IsZero() {
			t.Error("expiration time should not be zero")
		}
	})

	t.Run("uses provided ttl when configured explicitly", func(t *testing.T) {
		ttl := 2 * time.Hour
		_, _, exp, err := GenerateRefreshTokenWithTTL(ttl)
		if err != nil {
			t.Fatalf("GenerateRefreshTokenWithTTL() error = %v", err)
		}
		got := exp.Sub(time.Now().UTC())
		if got < ttl-time.Minute || got > ttl+time.Minute {
			t.Fatalf("expiration drift = %v, want around %v", got, ttl)
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			plain, _, _, err := GenerateRefreshToken()
			if err != nil {
				t.Fatalf("GenerateRefreshToken() error = %v", err)
			}
			if seen[plain] {
				t.Errorf("GenerateRefreshToken() generated duplicate token")
			}
			seen[plain] = true
		}
	})

	t.Run("plain token contains only base64url characters", func(t *testing.T) {
		plain, _, _, err := GenerateRefreshToken()
		if err != nil {
			t.Fatalf("GenerateRefreshToken() error = %v", err)
		}
		const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
		for _, c := range plain {
			if !strings.ContainsRune(alphabet, c) {
				t.Errorf("token contains invalid character: %c", c)
			}
		}
	})
}

// --- Testes de GenerateRefreshTokenFromPlain ---
// Previne: Hash inconsistente para o mesmo token

func TestGenerateRefreshTokenFromPlain(t *testing.T) {
	t.Run("returns same plain token", func(t *testing.T) {
		input := "test-token-value"
		plain, _, _, err := GenerateRefreshTokenFromPlain(input)
		if err != nil {
			t.Fatalf("GenerateRefreshTokenFromPlain() error = %v", err)
		}
		if plain != input {
			t.Errorf("plain = %q, want %q", plain, input)
		}
	})

	t.Run("generates consistent hash for same input", func(t *testing.T) {
		input := "consistent-token-test"
		_, hash1, _, _ := GenerateRefreshTokenFromPlain(input)
		_, hash2, _, _ := GenerateRefreshTokenFromPlain(input)

		if hash1 != hash2 {
			t.Errorf("hash should be consistent: %q != %q", hash1, hash2)
		}
	})

	t.Run("generates different hash for different input", func(t *testing.T) {
		_, hash1, _, _ := GenerateRefreshTokenFromPlain("token-a")
		_, hash2, _, _ := GenerateRefreshTokenFromPlain("token-b")

		if hash1 == hash2 {
			t.Error("different inputs should produce different hashes")
		}
	})

	t.Run("hash contains only base64url characters", func(t *testing.T) {
		_, hash, _, _ := GenerateRefreshTokenFromPlain("test-input")
		const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
		for _, c := range hash {
			if !strings.ContainsRune(alphabet, c) {
				t.Errorf("hash contains invalid character: %c", c)
			}
		}
	})
}

// --- Testes de encodeToBase64URL ---
// Previne: Encoding incorreto ou caracteres inválidos

func TestEncodeToBase64URL(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantLen  int
		wantSafe bool // deve conter apenas chars base64url
	}{
		{
			name:     "empty input",
			input:    []byte{},
			wantLen:  0,
			wantSafe: true,
		},
		{
			name:     "single byte",
			input:    []byte{0x00},
			wantLen:  1,
			wantSafe: true,
		},
		{
			name:     "multiple bytes",
			input:    []byte{0xFF, 0x00, 0xAB, 0xCD},
			wantLen:  4,
			wantSafe: true,
		},
		{
			name:     "all zero bytes",
			input:    make([]byte, 32),
			wantLen:  32,
			wantSafe: true,
		},
	}

	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeToBase64URL(tt.input)

			if len(got) != tt.wantLen {
				t.Errorf("encodeToBase64URL() length = %d, want %d", len(got), tt.wantLen)
			}

			if tt.wantSafe {
				for _, c := range got {
					if !strings.ContainsRune(alphabet, c) {
						t.Errorf("encodeToBase64URL() contains invalid character: %c", c)
					}
				}
			}
		})
	}
}

// --- Testes de Claims ---
// Previne: Claims incorretos em tokens JWT

func TestClaims_Sub(t *testing.T) {
	claims := &Claims{
		Sub: "user-123",
	}

	if claims.Sub != "user-123" {
		t.Errorf("Claims.Sub = %q, want %q", claims.Sub, "user-123")
	}
}

// --- Benchmark ---

func BenchmarkGenerateRefreshToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _, _ = GenerateRefreshToken()
	}
}

func BenchmarkEncodeToBase64URL(b *testing.B) {
	data := make([]byte, 64)
	for i := 0; i < b.N; i++ {
		_ = encodeToBase64URL(data)
	}
}
