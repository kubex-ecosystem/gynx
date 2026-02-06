// Package auth testa a lógica de domínio de autenticação.
//
// Este arquivo testa o método User.IsActive() que determina se um usuário
// está ativo no sistema. Previne regressões onde usuários inativos possam
// ser tratados como ativos (falha de segurança).
//
// Tipo: Teste unitário
package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "status active (lowercase) returns true",
			status: "active",
			want:   true,
		},
		{
			name:   "status ACTIVE (uppercase) returns true",
			status: "ACTIVE",
			want:   true,
		},
		{
			name:   "status inactive returns false",
			status: "inactive",
			want:   false,
		},
		{
			name:   "status suspended returns false",
			status: "suspended",
			want:   false,
		},
		{
			name:   "status pending returns false",
			status: "pending",
			want:   false,
		},
		{
			name:   "empty status returns false",
			status: "",
			want:   false,
		},
		{
			name:   "status with mixed case 'Active' returns false",
			status: "Active",
			want:   false,
		},
		{
			name:   "status with spaces returns false",
			status: " active ",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:        uuid.New(),
				Email:     "test@example.com",
				Name:      "Test User",
				Status:    tt.status,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			got := u.IsActive()
			if got != tt.want {
				t.Errorf("User.IsActive() = %v, want %v for status %q", got, tt.want, tt.status)
			}
		})
	}
}

// TestUser_IsActive_EdgeCases testa comportamentos de borda do método IsActive.
// Previne NPE e comportamentos inesperados com valores nulos ou vazios.
func TestUser_IsActive_WithMinimalFields(t *testing.T) {
	// User com apenas status preenchido - deve funcionar
	u := &User{Status: "active"}
	if !u.IsActive() {
		t.Error("User with minimal fields and status='active' should be active")
	}

	// User com status vazio
	u2 := &User{}
	if u2.IsActive() {
		t.Error("User with empty status should not be active")
	}
}
