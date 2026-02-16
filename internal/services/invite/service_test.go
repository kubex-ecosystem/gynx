// Package invite testa a lógica de negócio do serviço de convites.
//
// Estes testes validam:
// - Validação de input (validateCreateReq)
// - Cálculo de expiração (computeExpiry)
// - Geração de token seguro (generateToken)
// - Validação de status e expiração de convites (validateInvite)
// - Funções helper de conversão (safeString, ptrVal, ptrOrNil, toDTO)
//
// Tipo: Testes unitários (sem dependências externas)
package invite

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	api "github.com/kubex-ecosystem/gnyx/internal/api/invite"
	auth "github.com/kubex-ecosystem/gnyx/internal/domain/auth"
	domain "github.com/kubex-ecosystem/gnyx/internal/domain/invites"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/internal/services/mailer"
)

// --- Stubs simples para testes ---

type stubMailer struct {
	lastMessage *mailer.EmailMessage
	sendErr     error
}

func (s *stubMailer) Send(msg *mailer.EmailMessage) error {
	s.lastMessage = msg
	return s.sendErr
}

type stubTemplateEngine struct {
	subject string
	html    string
	err     error
}

func (s *stubTemplateEngine) Render(templateName string, data mailer.TemplateData) (string, string, error) {
	return s.subject, s.html, s.err
}

type stubAdapter struct{}

// stubUserRepo implementa userstore.UserRepository para testes
type stubUserRepo struct{}

func (s *stubUserRepo) FindByEmail(ctx context.Context, email string) (*auth.User, error) {
	return nil, nil
}
func (s *stubUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	return nil, nil
}
func (s *stubUserRepo) Create(ctx context.Context, user *auth.User) error {
	return nil
}
func (s *stubUserRepo) ListMemberships(ctx context.Context, userID uuid.UUID) ([]auth.Membership, error) {
	return nil, nil
}

func (s *stubAdapter) GetByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	return nil, dsclient.ErrNotFound
}
func (s *stubAdapter) GetByID(ctx context.Context, id string, invType domain.InvitationType) (*domain.Invitation, error) {
	return nil, dsclient.ErrNotFound
}
func (s *stubAdapter) CreatePartner(ctx context.Context, input *domain.CreatePartnerInvitationInput) (*domain.Invitation, error) {
	return nil, nil
}
func (s *stubAdapter) CreateInternal(ctx context.Context, input *domain.CreateInternalInvitationInput) (*domain.Invitation, error) {
	return nil, nil
}
func (s *stubAdapter) Update(ctx context.Context, id string, invType domain.InvitationType, input *domain.UpdateInvitationInput) (*domain.Invitation, error) {
	return nil, nil
}
func (s *stubAdapter) Revoke(ctx context.Context, id string, invType domain.InvitationType) error {
	return nil
}
func (s *stubAdapter) Accept(ctx context.Context, token string) (*domain.Invitation, error) {
	return nil, nil
}
func (s *stubAdapter) Delete(ctx context.Context, id string, invType domain.InvitationType) error {
	return nil
}
func (s *stubAdapter) List(ctx context.Context, filters *domain.InvitationFilters) (*domain.PaginatedInvitations, error) {
	return &domain.PaginatedInvitations{Data: []*domain.Invitation{}}, nil
}

// --- Testes de validateCreateReq ---
// Previne: Convites criados com dados obrigatórios ausentes

func TestValidateCreateReq(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name    string
		req     api.CreateInviteReq
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with all required fields",
			req: api.CreateInviteReq{
				Email:     "user@example.com",
				Role:      "partner",
				TenantID:  "550e8400-e29b-41d4-a716-446655440000",
				InvitedBy: "admin@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			req: api.CreateInviteReq{
				Role:      "partner",
				TenantID:  "550e8400-e29b-41d4-a716-446655440000",
				InvitedBy: "admin@example.com",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "empty email (whitespace only)",
			req: api.CreateInviteReq{
				Email:     "   ",
				Role:      "partner",
				TenantID:  "550e8400-e29b-41d4-a716-446655440000",
				InvitedBy: "admin@example.com",
			},
			wantErr: true,
			errMsg:  "email is required",
		},
		{
			name: "missing role",
			req: api.CreateInviteReq{
				Email:     "user@example.com",
				TenantID:  "550e8400-e29b-41d4-a716-446655440000",
				InvitedBy: "admin@example.com",
			},
			wantErr: true,
			errMsg:  "role is required",
		},
		{
			name: "missing tenant_id",
			req: api.CreateInviteReq{
				Email:     "user@example.com",
				Role:      "partner",
				InvitedBy: "admin@example.com",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "missing invited_by",
			req: api.CreateInviteReq{
				Email:    "user@example.com",
				Role:     "partner",
				TenantID: "550e8400-e29b-41d4-a716-446655440000",
			},
			wantErr: true,
			errMsg:  "invited_by is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateCreateReq(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCreateReq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateCreateReq() error = %q, want error containing %q", err.Error(), tt.errMsg)
			}
		})
	}
}

// --- Testes de computeExpiry ---
// Previne: Convites com datas de expiração incorretas

func TestComputeExpiry(t *testing.T) {
	defaultTTL := 7 * 24 * time.Hour
	svc := &Service{defaultTTL: defaultTTL}

	t.Run("uses custom days when provided", func(t *testing.T) {
		days := 14
		before := time.Now().UTC()
		result := svc.computeExpiry(days)
		after := time.Now().UTC()

		expectedMin := before.Add(time.Duration(days) * 24 * time.Hour)
		expectedMax := after.Add(time.Duration(days) * 24 * time.Hour)

		if result.Before(expectedMin) || result.After(expectedMax) {
			t.Errorf("computeExpiry(%d) = %v, want between %v and %v", days, result, expectedMin, expectedMax)
		}
	})

	t.Run("uses default TTL when days is 0", func(t *testing.T) {
		before := time.Now().UTC()
		result := svc.computeExpiry(0)
		after := time.Now().UTC()

		expectedMin := before.Add(defaultTTL)
		expectedMax := after.Add(defaultTTL)

		if result.Before(expectedMin) || result.After(expectedMax) {
			t.Errorf("computeExpiry(0) = %v, want between %v and %v", result, expectedMin, expectedMax)
		}
	})

	t.Run("uses default TTL when days is negative", func(t *testing.T) {
		before := time.Now().UTC()
		result := svc.computeExpiry(-5)
		after := time.Now().UTC()

		expectedMin := before.Add(defaultTTL)
		expectedMax := after.Add(defaultTTL)

		if result.Before(expectedMin) || result.After(expectedMax) {
			t.Errorf("computeExpiry(-5) = %v, want between %v and %v", result, expectedMin, expectedMax)
		}
	})

	t.Run("returns UTC time", func(t *testing.T) {
		result := svc.computeExpiry(1)
		if result.Location() != time.UTC {
			t.Errorf("computeExpiry() returned non-UTC time: %v", result.Location())
		}
	})
}

// --- Testes de generateToken ---
// Previne: Tokens previsíveis ou duplicados (falha de segurança)

func TestGenerateToken(t *testing.T) {
	svc := &Service{}

	t.Run("generates 64-character hex string", func(t *testing.T) {
		token, err := svc.generateToken()
		if err != nil {
			t.Fatalf("generateToken() error = %v", err)
		}
		if len(token) != 64 {
			t.Errorf("generateToken() length = %d, want 64", len(token))
		}
		// Verifica se é hex válido
		for _, c := range token {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("generateToken() contains invalid hex character: %c", c)
			}
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			token, err := svc.generateToken()
			if err != nil {
				t.Fatalf("generateToken() error = %v", err)
			}
			if seen[token] {
				t.Errorf("generateToken() generated duplicate token: %s", token)
			}
			seen[token] = true
		}
	})
}

// --- Testes de validateInvite ---
// Previne: Aceitar convites expirados ou já processados

func TestValidateInvite(t *testing.T) {
	tests := []struct {
		name    string
		inv     *domain.Invitation
		wantErr error
	}{
		{
			name:    "nil invitation returns ErrNotFound",
			inv:     nil,
			wantErr: dsclient.ErrNotFound,
		},
		{
			name: "accepted invitation returns ErrInvalidStatus",
			inv: &domain.Invitation{
				Invitation: dsclient.Invitation{
					Status:    dsclient.StatusAccepted,
					ExpiresAt: time.Now().Add(time.Hour),
				},
			},
			wantErr: dsclient.ErrInvalidStatus,
		},
		{
			name: "revoked invitation returns ErrInvalidStatus",
			inv: &domain.Invitation{
				Invitation: dsclient.Invitation{
					Status:    dsclient.StatusRevoked,
					ExpiresAt: time.Now().Add(time.Hour),
				},
			},
			wantErr: dsclient.ErrInvalidStatus,
		},
		{
			name: "expired invitation returns ErrExpired",
			inv: &domain.Invitation{
				Invitation: dsclient.Invitation{
					Status:    dsclient.StatusPending,
					ExpiresAt: time.Now().Add(-time.Hour), // já expirou
				},
			},
			wantErr: dsclient.ErrExpired,
		},
		{
			name: "valid pending invitation returns nil",
			inv: &domain.Invitation{
				Invitation: dsclient.Invitation{
					Status:    dsclient.StatusPending,
					ExpiresAt: time.Now().Add(time.Hour), // ainda válido
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInvite(tt.inv)
			if err != tt.wantErr {
				t.Errorf("validateInvite() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- Testes de safeString ---
// Previne: Valores vazios onde esperamos fallback

func TestSafeString(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback string
		want     string
	}{
		{
			name:     "returns value when not empty",
			value:    "hello",
			fallback: "world",
			want:     "hello",
		},
		{
			name:     "returns fallback when value is empty",
			value:    "",
			fallback: "fallback",
			want:     "fallback",
		},
		{
			name:     "returns fallback when value is whitespace",
			value:    "   ",
			fallback: "fallback",
			want:     "fallback",
		},
		{
			name:     "trims value",
			value:    "  hello  ",
			fallback: "world",
			want:     "hello",
		},
		{
			name:     "trims fallback",
			value:    "",
			fallback: "  world  ",
			want:     "world",
		},
		{
			name:     "returns empty when both are empty",
			value:    "",
			fallback: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeString(tt.value, tt.fallback)
			if got != tt.want {
				t.Errorf("safeString(%q, %q) = %q, want %q", tt.value, tt.fallback, got, tt.want)
			}
		})
	}
}

// --- Testes de ptrVal ---
// Previne: Panic ao acessar ponteiro nulo

func TestPtrVal(t *testing.T) {
	tests := []struct {
		name  string
		value *string
		want  string
	}{
		{
			name:  "returns empty string for nil pointer",
			value: nil,
			want:  "",
		},
		{
			name:  "returns value for non-nil pointer",
			value: ptr("hello"),
			want:  "hello",
		},
		{
			name:  "returns empty string for pointer to empty string",
			value: ptr(""),
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ptrVal(tt.value)
			if got != tt.want {
				t.Errorf("ptrVal() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- Testes de ptrOrNil ---
// Previne: Ponteiros para strings vazias onde nil é esperado

func TestPtrOrNil(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  *string
	}{
		{
			name:  "returns nil for empty string",
			value: "",
			want:  nil,
		},
		{
			name:  "returns nil for whitespace only",
			value: "   ",
			want:  nil,
		},
		{
			name:  "returns pointer for non-empty string",
			value: "hello",
			want:  ptr("hello"),
		},
		{
			name:  "trims and returns pointer",
			value: "  hello  ",
			want:  ptr("hello"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ptrOrNil(tt.value)
			if tt.want == nil {
				if got != nil {
					t.Errorf("ptrOrNil(%q) = %v, want nil", tt.value, *got)
				}
			} else {
				if got == nil {
					t.Errorf("ptrOrNil(%q) = nil, want %q", tt.value, *tt.want)
				} else if *got != *tt.want {
					t.Errorf("ptrOrNil(%q) = %q, want %q", tt.value, *got, *tt.want)
				}
			}
		})
	}
}

// --- Testes de toDTO ---
// Previne: Campos mapeados incorretamente entre camadas

func TestToDTO(t *testing.T) {
	t.Run("returns nil for nil invitation", func(t *testing.T) {
		dto := toDTO(nil)
		if dto != nil {
			t.Error("toDTO(nil) should return nil")
		}
	})

	t.Run("maps all fields correctly", func(t *testing.T) {
		teamID := "team-123"
		expiresAt := time.Now().Add(24 * time.Hour)

		inv := &domain.Invitation{
			Invitation: dsclient.Invitation{
				ID:        "inv-123",
				Email:     "test@example.com",
				Role:      "partner",
				TenantID:  "tenant-456",
				TeamID:    &teamID,
				Status:    dsclient.StatusPending,
				Type:      dsclient.TypePartner,
				ExpiresAt: expiresAt,
			},
		}

		dto := toDTO(inv)

		if dto.ID != inv.ID {
			t.Errorf("ID = %q, want %q", dto.ID, inv.ID)
		}
		if dto.Email != inv.Email {
			t.Errorf("Email = %q, want %q", dto.Email, inv.Email)
		}
		if dto.Role != inv.Role {
			t.Errorf("Role = %q, want %q", dto.Role, inv.Role)
		}
		if dto.TenantID != inv.TenantID {
			t.Errorf("TenantID = %q, want %q", dto.TenantID, inv.TenantID)
		}
		if dto.TeamID != teamID {
			t.Errorf("TeamID = %q, want %q", dto.TeamID, teamID)
		}
		if dto.Status != string(dsclient.StatusPending) {
			t.Errorf("Status = %q, want %q", dto.Status, string(dsclient.StatusPending))
		}
		if dto.Type != string(dsclient.TypePartner) {
			t.Errorf("Type = %q, want %q", dto.Type, string(dsclient.TypePartner))
		}
		expectedExpires := expiresAt.Format(time.RFC3339)
		if dto.ExpiresAt != expectedExpires {
			t.Errorf("ExpiresAt = %q, want %q", dto.ExpiresAt, expectedExpires)
		}
	})

	t.Run("handles nil TeamID", func(t *testing.T) {
		inv := &domain.Invitation{
			Invitation: dsclient.Invitation{
				ID:        "inv-123",
				Email:     "test@example.com",
				Role:      "partner",
				TenantID:  "tenant-456",
				TeamID:    nil,
				Status:    dsclient.StatusPending,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
		}

		dto := toDTO(inv)

		if dto.TeamID != "" {
			t.Errorf("TeamID = %q, want empty string for nil TeamID", dto.TeamID)
		}
	})
}

// --- Testes de NewService ---
// Previne: Serviço criado com configuração inválida

func TestNewService(t *testing.T) {
	t.Run("returns error when adapter is nil", func(t *testing.T) {
		cfg := Config{
			Adapter:   nil,
			Mailer:    &stubMailer{},
			Templates: &stubTemplateEngine{},
		}
		_, err := NewService(cfg)
		if err == nil || !strings.Contains(err.Error(), "adapter is required") {
			t.Errorf("NewService() error = %v, want error about adapter", err)
		}
	})

	t.Run("returns error when mailer is nil", func(t *testing.T) {
		cfg := Config{
			Adapter:   &stubAdapter{},
			Mailer:    nil,
			Templates: &stubTemplateEngine{},
		}
		_, err := NewService(cfg)
		if err == nil || !strings.Contains(err.Error(), "mailer is required") {
			t.Errorf("NewService() error = %v, want error about mailer", err)
		}
	})

	t.Run("returns error when templates is nil", func(t *testing.T) {
		cfg := Config{
			Adapter:   &stubAdapter{},
			Mailer:    &stubMailer{},
			Templates: nil,
		}
		_, err := NewService(cfg)
		if err == nil || !strings.Contains(err.Error(), "template engine is required") {
			t.Errorf("NewService() error = %v, want error about templates", err)
		}
	})

	// Nota: Este teste requer configuração do UserRepository que depende de
	// ambiente externo. Testamos apenas os valores default do serviço quando
	// todos os campos obrigatórios são fornecidos.
	t.Run("uses provided config values", func(t *testing.T) {
		cfg := Config{
			Adapter:     &stubAdapter{},
			Mailer:      &stubMailer{},
			Templates:   &stubTemplateEngine{},
			BaseURL:     "https://custom.api.com",
			SenderName:  "Custom Sender",
			SenderEmail: "custom@example.com",
			CompanyName: "Custom Company",
			DefaultTTL:  48 * time.Hour,
			UserRepo:    &stubUserRepo{}, // mock para evitar init real
		}
		svc, err := NewService(cfg)
		if err != nil {
			t.Fatalf("NewService() error = %v", err)
		}

		if svc.baseURL != "https://custom.api.com" {
			t.Errorf("baseURL = %q, want custom", svc.baseURL)
		}
		if svc.senderName != "Custom Sender" {
			t.Errorf("senderName = %q, want custom", svc.senderName)
		}
		if svc.senderEmail != "custom@example.com" {
			t.Errorf("senderEmail = %q, want custom", svc.senderEmail)
		}
		if svc.companyName != "Custom Company" {
			t.Errorf("companyName = %q, want custom", svc.companyName)
		}
		if svc.defaultTTL != 48*time.Hour {
			t.Errorf("defaultTTL = %v, want 48h", svc.defaultTTL)
		}
	})
}

// --- Testes de inviteURL ---
// Previne: URLs malformadas nos emails de convite

func TestInviteURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		token   string
		want    string
	}{
		{
			name:    "basic URL construction",
			baseURL: "https://api.kubex.world",
			token:   "abc123",
			want:    "https://api.kubex.world/invite/abc123",
		},
		{
			name:    "removes trailing slash from baseURL",
			baseURL: "https://api.kubex.world/",
			token:   "abc123",
			want:    "https://api.kubex.world/invite/abc123",
		},
		{
			name:    "handles multiple trailing slashes",
			baseURL: "https://api.kubex.world///",
			token:   "abc123",
			want:    "https://api.kubex.world/invite/abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{baseURL: tt.baseURL}
			got := svc.inviteURL(tt.token)
			if got != tt.want {
				t.Errorf("inviteURL(%q) = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}

// Benchmark para generateToken - garante performance aceitável
func BenchmarkGenerateToken(b *testing.B) {
	svc := &Service{}
	for i := 0; i < b.N; i++ {
		_, _ = svc.generateToken()
	}
}
