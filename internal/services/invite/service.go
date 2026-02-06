// Package invite implementa o serviço de convites por e-mail.
package invite

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	api "github.com/kubex-ecosystem/gnyx/internal/api/invite"
	auth "github.com/kubex-ecosystem/gnyx/internal/domain/auth"
	domain "github.com/kubex-ecosystem/gnyx/internal/domain/invites"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore"
	userstore "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/user_store"
	"github.com/kubex-ecosystem/gnyx/internal/services/mailer"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
	"golang.org/x/crypto/bcrypt"
)

// MailSender define o contrato mínimo para envio de e-mails.
type MailSender interface {
	Send(msg *mailer.EmailMessage) error
}

// TemplateEngine define o contrato para renderização de templates de e-mail.
type TemplateEngine interface {
	Render(templateName string, data mailer.TemplateData) (subject string, html string, err error)
}

// Config agrega dependências do serviço de convites.
type Config struct {
	Adapter     domain.Adapter           `json:"adapter" yaml:"adapter" xml:"adapter" toml:"adapter"`
	Mailer      MailSender               `json:"mailer" yaml:"mailer" xml:"mailer" toml:"mailer"`
	Templates   TemplateEngine           `json:"templates" yaml:"templates" xml:"templates" toml:"templates"`
	BaseURL     string                   `json:"base_url" yaml:"base_url" xml:"base_url" toml:"base_url"`
	SenderName  string                   `json:"sender_name" yaml:"sender_name" xml:"sender_name" toml:"sender_name"`
	SenderEmail string                   `json:"sender_email" yaml:"sender_email" xml:"sender_email" toml:"sender_email"`
	CompanyName string                   `json:"company_name" yaml:"company_name" xml:"company_name" toml:"company_name"`
	DefaultTTL  time.Duration            `json:"default_ttl" yaml:"default_ttl" xml:"default_ttl" toml:"default_ttl"`
	UserRepo    userstore.UserRepository `json:"user_repo" yaml:"user_repo" xml:"user_repo" toml:"user_repo"`
}

// Service implementa api.Service usando o banco do DataService.
type Service struct {
	repo        domain.Adapter           `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	mailer      MailSender               `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	templates   TemplateEngine           `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	baseURL     string                   `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	senderName  string                   `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	senderEmail string                   `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	companyName string                   `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	defaultTTL  time.Duration            `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	userRepo    userstore.UserRepository `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
}

func NewAdapter(store dsclient.InviteStore) (domain.Adapter, error) {
	if store == nil {
		return nil, gl.Errorf("store is required")
	}
	return &dsAdapter{store: store}, nil
}

// NewService cria uma nova instância do serviço de convites.
func NewService(cfg Config) (*Service, error) {
	if cfg.Adapter == nil {
		return nil, errors.New("invite adapter is required")
	}
	if cfg.Mailer == nil {
		return nil, errors.New("mailer is required")
	}
	if cfg.Templates == nil {
		return nil, errors.New("template engine is required")
	}
	if cfg.DefaultTTL <= 0 {
		cfg.DefaultTTL = 7 * 24 * time.Hour
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.gnyx.app"
	}
	if cfg.SenderName == "" {
		cfg.SenderName = "Equipe Kubex"
	}
	if cfg.SenderEmail == "" {
		cfg.SenderEmail = "convites@gnyx.app"
	}
	if cfg.CompanyName == "" {
		cfg.CompanyName = "GNyxM"
	}
	if cfg.UserRepo == nil {
		repo, err := userstore.NewUserRepository()
		if err != nil {
			return nil, gl.Errorf("failed to init user repository: %v", err)
		}
		cfg.UserRepo = repo
	}

	return &Service{
		repo:        cfg.Adapter,
		mailer:      cfg.Mailer,
		templates:   cfg.Templates,
		baseURL:     cfg.BaseURL,
		senderName:  cfg.SenderName,
		senderEmail: cfg.SenderEmail,
		companyName: cfg.CompanyName,
		defaultTTL:  cfg.DefaultTTL,
		userRepo:    cfg.UserRepo,
	}, nil
}

// CreateInvite cria e envia um novo convite.
func (s *Service) CreateInvite(ctx context.Context, req api.CreateInviteReq) (*api.InviteDTO, error) {
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = strings.TrimSpace(req.RoleCode)
	}
	if role == "" {
		role = strings.TrimSpace(req.RoleID)
	}
	if role == "" {
		role = "viewer" // default role fallback
	}
	req.Role = role

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = strings.ToTitleSpecial(unicode.CaseRanges, strings.Split(req.Email, "@")[0])
	}
	req.Name = name

	if err := s.validateCreateReq(req); err != nil {
		return nil, err
	}

	// Idempotência: reutiliza convite pendente existente para o mesmo email/tenant/tipo.
	existing, err := s.findPendingInvite(ctx, req)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Refresca expiração e reenvia email
		newExpiry := s.computeExpiry(req.ExpiresInDays)
		if _, err := s.repo.Update(ctx, existing.ID, existing.Type, &domain.UpdateInvitationInput{ExpiresAt: &newExpiry}); err == nil {
			existing.ExpiresAt = newExpiry
		}
		if err := s.sendInviteEmail(existing, existing.Token, req); err != nil {
			return nil, err
		}
		dto := toDTO(existing)
		dto.Token = existing.Token
		return dto, nil
	}

	token, err := s.generateToken()
	if err != nil {
		return nil, gl.Errorf("failed to generate invite token: %v", err)
	}

	req.ExpiresAt = s.computeExpiry(req.ExpiresInDays)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var (
		inv *domain.Invitation
	)

	if strings.TrimSpace(req.TeamID) != "" {
		inv, err = s.repo.CreateInternal(ctx, &domain.CreateInternalInvitationInput{
			Token:        token,
			InviteeEmail: req.Email,
			InviteeName:  &req.Name,
			Role:         req.Role,
			TeamID:       ptrOrNil(req.TeamID),
			TenantID:     req.TenantID,
			InvitedBy:    req.InvitedBy,
			ExpiresAt:    &req.ExpiresAt,
		})
	} else {
		inv, err = s.repo.CreatePartner(ctx, &domain.CreatePartnerInvitationInput{
			Token:        token,
			PartnerEmail: req.Email,
			PartnerName:  ptrOrNil(req.Name),
			CompanyName:  ptrOrNil(req.CompanyName),
			Role:         req.Role,
			TenantID:     req.TenantID,
			InvitedBy:    req.InvitedBy,
			ExpiresAt:    &req.ExpiresAt,
		})
	}

	if err != nil {
		return nil, err
	}

	if err := s.sendInviteEmail(inv, token, req); err != nil {
		// rollback best-effort
		if delErr := s.repo.Delete(ctx, inv.ID, inv.Type); delErr != nil {
			gl.Log("error", fmt.Sprintf("failed to cleanup invite %s after mail error: %v", inv.ID, delErr))
		}
		return nil, err
	}

	dto := toDTO(inv)
	dto.Token = token
	return dto, nil
}

// ValidateToken retorna os dados do convite sem expor o token novamente.
func (s *Service) ValidateToken(ctx context.Context, token string) (*api.InviteDTO, error) {
	inv, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return toDTO(inv), nil
}

// AcceptInvite marca o convite como aceito.
func (s *Service) AcceptInvite(ctx context.Context, token string, req api.AcceptInviteReq) (*api.AcceptResult, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("token is required")
	}

	inv, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	userID, err := s.ensureUser(ctx, inv, req)
	if err != nil {
		return nil, err
	}

	if err := s.ensureMembership(ctx, userID, inv); err != nil {
		return nil, err
	}

	inv, err = s.repo.Accept(ctx, token)
	if err != nil {
		return nil, err
	}

	result := &api.AcceptResult{
		UserID:     userID,
		TenantID:   inv.TenantID,
		Membership: inv.Role,
	}

	_ = req // placeholder para campos futuros
	return result, nil
}

// ListInvites lista convites para o FE com filtros básicos.
func (s *Service) ListInvites(ctx context.Context, filters api.InviteListFilters) (*api.InviteListResponse, error) {
	if s.repo == nil {
		return nil, errors.New("invite repository not configured")
	}

	// default page/limit
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}

	invType := domain.TypePartner
	if strings.ToLower(strings.TrimSpace(filters.Type)) == string(domain.TypeInternal) {
		invType = domain.TypeInternal
	}

	var status *domain.InvitationStatus
	if st := strings.TrimSpace(filters.Status); st != "" {
		s := domain.InvitationStatus(strings.ToLower(st))
		status = &s
	}

	var emailPtr *string
	if e := strings.TrimSpace(filters.Email); e != "" {
		emailPtr = &e
	}
	var tenantPtr *string
	if t := strings.TrimSpace(filters.TenantID); t != "" {
		tenantPtr = &t
	}

	res, err := s.repo.List(ctx, &domain.InvitationFilters{
		Type:     &invType,
		Status:   status,
		Email:    emailPtr,
		TenantID: tenantPtr,
		Page:     filters.Page,
		Limit:    filters.Limit,
	})
	if err != nil {
		return nil, err
	}

	out := &api.InviteListResponse{
		Data:       make([]*api.InviteDTO, 0, len(res.Data)),
		Total:      res.Total,
		Page:       res.Page,
		Limit:      res.Limit,
		TotalPages: res.TotalPages,
	}
	for _, inv := range res.Data {
		out.Data = append(out.Data, toDTO(inv))
	}
	return out, nil
}

// Helpers ----------------------------------------------------------------

func (s *Service) validateCreateReq(req api.CreateInviteReq) error {
	if strings.TrimSpace(req.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(req.Role) == "" {
		return errors.New("role is required")
	}
	if strings.TrimSpace(req.TenantID) == "" {
		return errors.New("tenant_id is required")
	}
	if strings.TrimSpace(req.InvitedBy) == "" {
		return errors.New("invited_by is required")
	}
	return nil
}

func (s *Service) computeExpiry(days int) time.Time {
	if days > 0 {
		return time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
	}
	return time.Now().UTC().Add(s.defaultTTL)
}

func (s *Service) generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *Service) sendInviteEmail(inv *domain.Invitation, token string, req api.CreateInviteReq) error {
	inviteURL := s.inviteURL(token)

	templateEngine := kbxGet.ValueOrIf(s.templates == nil, nil, s.templates)
	if templateEngine == nil {
		te, err := mailer.GetDefaultTemplateLoader()
		if err != nil {
			gl.Warnf("failed to init default template loader: %v", err)
		} else {
			templateEngine = te
		}
	}
	s.templates = templateEngine

	data := mailer.TemplateData{
		Type:          "Convite GNyxM",
		Email:         inv.Email,
		RecipientName: strings.TrimSpace(kbxGet.ValueOrIf(strings.TrimSpace(inv.Name) != "", inv.Name, inv.Email)),
		CompanyName:   s.companyName,
		SiteURL:       s.baseURL,
		ActionURL:     inviteURL,
		ActionLabel:   "Aceitar convite",
		Data: map[string]any{
			"role_name":       inv.Role,
			"invited_by_name": req.InvitedBy,
			"expires_at":      inv.ExpiresAt.Format(time.RFC3339),
			"invitation_url":  inviteURL,
			"invite_link":     inviteURL,
			"name":            strings.TrimSpace(kbxGet.ValueOrIf(strings.TrimSpace(inv.Name) != "", inv.Name, inv.Email)),
		},
	}

	subject := "Convite GNyxM"
	html := ""
	var err error
	if s.templates != nil {
		if subject, html, err = s.templates.Render("user_invited", data); err != nil {
			gl.Warnf("failed to render invite template with configured loader: %v", err)
			// tenta loader padrão embedado antes de cair no texto puro
			if def, defErr := mailer.GetDefaultTemplateLoader(); defErr == nil {
				if subject, html, err = def.Render("user_invited", data); err != nil {
					gl.Warnf("failed to render invite template with default loader: %v", err)
				}
			} else {
				gl.Warnf("failed to init default template loader: %v", defErr)
			}
		}
	} else {
		gl.Warn("invite templates loader is nil; sending plain text fallback")
	}

	message := &mailer.EmailMessage{
		To:      []string{inv.Email},
		Subject: subject,
		HTML:    kbxGet.ValueOrIf(strings.TrimSpace(html) != "", html, s.plainTextFallback(inv, inviteURL)),
		Text:    kbxGet.ValueOrIf(strings.TrimSpace(html) != "", "", s.plainTextFallback(inv, inviteURL)),
		From:    s.senderEmail,
		Name:    s.senderName,
	}

	return s.mailer.Send(message)
}

func (s *Service) inviteURL(token string) string {
	base := strings.TrimRight(s.baseURL, "/")
	return fmt.Sprintf("%s/invite/%s", base, token)
}

func (s *Service) plainTextFallback(inv *domain.Invitation, url string) string {
	return fmt.Sprintf("Olá %s,\n\nVocê foi convidado para acessar a GNyxM como %s.\nUse o link a seguir para aceitar o convite: %s\n\nConvite válido até %s.\n",
		safeString(inv.Name, inv.Email),
		inv.Role,
		url,
		inv.ExpiresAt.Format(time.RFC1123))
}

// findPendingInvite busca um convite pendente para o mesmo email/tenant/tipo para evitar duplicidade.
func (s *Service) findPendingInvite(ctx context.Context, req api.CreateInviteReq) (*domain.Invitation, error) {
	if s.repo == nil {
		return nil, errors.New("invite repository not configured")
	}
	invType := domain.TypePartner
	if strings.TrimSpace(req.TeamID) != "" {
		invType = domain.TypeInternal
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	tenant := strings.TrimSpace(req.TenantID)
	status := domain.StatusPending

	res, err := s.repo.List(ctx, &domain.InvitationFilters{
		Type:     &invType,
		Status:   &status,
		Email:    ptr(email),
		TenantID: ptr(tenant),
		Page:     1,
		Limit:    1,
	})
	if err != nil {
		return nil, err
	}
	if res != nil && len(res.Data) > 0 {
		return res.Data[0], nil
	}
	return nil, nil
}

func ptr[T any](v T) *T { return &v }

func toDTO(inv *domain.Invitation) *api.InviteDTO {
	if inv == nil {
		return nil
	}
	return &api.InviteDTO{
		ID:        inv.ID,
		Email:     inv.Email,
		Role:      inv.Role,
		TenantID:  inv.TenantID,
		TeamID:    ptrVal(inv.TeamID),
		Status:    string(inv.Status),
		ExpiresAt: inv.ExpiresAt.Format(time.RFC3339),
		Type:      string(inv.Type),
	}
}

func safeString(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(fallback)
}

func ptrVal(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func ptrOrNil(value string) *string {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	return &v
}

func (s *Service) ensureUser(ctx context.Context, inv *domain.Invitation, req api.AcceptInviteReq) (string, error) {
	if s.userRepo == nil {
		return "", errors.New("user repository not configured")
	}

	email := strings.ToLower(strings.TrimSpace(inv.Email))
	if email == "" {
		return "", errors.New("invitation email is empty")
	}

	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		return existing.ID.String(), nil
	}
	if err != nil && !errors.Is(err, userstore.ErrUserNotFound) {
		return "", err
	}

	password := strings.TrimSpace(req.Password)
	forceReset := false
	if password == "" {
		password, err = s.generateToken()
		if err != nil {
			return "", err
		}
		forceReset = true
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &auth.User{
		Email:              email,
		Name:               safeString(req.Name, safeString(inv.Name, email)),
		LastName:           strings.TrimSpace(req.LastName),
		PasswordHash:       string(hash),
		Status:             "active",
		ForcePasswordReset: forceReset,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", err
	}

	return user.ID.String(), nil
}

func (s *Service) ensureMembership(ctx context.Context, userID string, inv *domain.Invitation) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("user_id is required for membership")
	}
	if inv == nil {
		return errors.New("invitation is required for membership")
	}

	conn, err := datastore.Connection(ctx)
	if err != nil {
		return err
	}
	exec, err := datastore.GetPGExecutor(ctx, conn)
	if err != nil {
		return err
	}

	var roleID uuid.UUID
	if err := exec.QueryRow(ctx, `SELECT id FROM role WHERE code = $1`, inv.Role).Scan(&roleID); err != nil {
		if err := exec.QueryRow(ctx, `SELECT id FROM role WHERE code = $1`, "viewer").Scan(&roleID); err != nil {
			return gl.Errorf("failed to resolve role '%s' and fallback 'viewer': %v", inv.Role, err)
		}
	}

	if _, err := exec.Exec(ctx, `
		INSERT INTO tenant_membership (user_id, tenant_id, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, now(), now())
		ON CONFLICT (user_id, tenant_id)
		DO UPDATE SET role_id = EXCLUDED.role_id, is_active = EXCLUDED.is_active, updated_at = now()
	`, userID, inv.TenantID, roleID); err != nil {
		return gl.Errorf("failed to upsert tenant membership: %v", err)
	}

	if inv.TeamID != nil && strings.TrimSpace(*inv.TeamID) != "" {
		if _, err := exec.Exec(ctx, `
			INSERT INTO team_membership (user_id, team_id, role_id, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, true, now(), now())
			ON CONFLICT (user_id, team_id)
			DO UPDATE SET role_id = EXCLUDED.role_id, is_active = EXCLUDED.is_active, updated_at = now()
		`, userID, *inv.TeamID, roleID); err != nil {
			return gl.Errorf("failed to upsert team membership: %v", err)
		}
	}

	return nil
}

// dsAdapter adapta o InviteStore do DS para o contrato domain.Adapter.
type dsAdapter struct {
	store dsclient.InviteStore
}

func (a *dsAdapter) GetByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	inv, err := a.store.GetByToken(ctx, strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, dsclient.ErrNotFound
	}
	domainInv := toDomain(inv)
	if err := validateInvite(domainInv); err != nil {
		return nil, err
	}
	return domainInv, nil
}

func (a *dsAdapter) GetByID(ctx context.Context, id string, invType domain.InvitationType) (*domain.Invitation, error) {
	dsType := toDSType(invType)
	inv, err := a.store.GetByID(ctx, id, dsType)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, dsclient.ErrNotFound
	}
	return toDomain(inv), nil
}

func (a *dsAdapter) CreatePartner(ctx context.Context, input *domain.CreatePartnerInvitationInput) (*domain.Invitation, error) {
	if input == nil {
		return nil, gl.Errorf("input is required")
	}
	dsInput := &dsclient.CreateInvitationInput{
		Type:      dsclient.TypePartner,
		Name:      safeString(ptrVal(input.PartnerName), input.PartnerEmail),
		Email:     strings.ToLower(strings.TrimSpace(input.PartnerEmail)),
		Role:      input.Role,
		Token:     input.Token,
		TenantID:  input.TenantID,
		TeamID:    nil,
		InvitedBy: input.InvitedBy,
		ExpiresAt: normalizeTime(input.ExpiresAt),
	}
	inv, err := a.store.Create(ctx, dsInput)
	if err != nil {
		return nil, err
	}
	return toDomain(inv), nil
}

func (a *dsAdapter) CreateInternal(ctx context.Context, input *domain.CreateInternalInvitationInput) (*domain.Invitation, error) {
	if input == nil {
		return nil, gl.Errorf("input is required")
	}
	dsInput := &dsclient.CreateInvitationInput{
		Type:      dsclient.TypeInternal,
		Name:      safeString(ptrVal(input.InviteeName), input.InviteeEmail),
		Email:     strings.ToLower(strings.TrimSpace(input.InviteeEmail)),
		Role:      input.Role,
		Token:     input.Token,
		TenantID:  input.TenantID,
		TeamID:    input.TeamID,
		InvitedBy: input.InvitedBy,
		ExpiresAt: normalizeTime(input.ExpiresAt),
	}
	inv, err := a.store.Create(ctx, dsInput)
	if err != nil {
		return nil, err
	}
	return toDomain(inv), nil
}

func (a *dsAdapter) Update(ctx context.Context, id string, invType domain.InvitationType, input *domain.UpdateInvitationInput) (*domain.Invitation, error) {
	if input == nil {
		return nil, gl.Errorf("update input is required")
	}
	dsInput := &dsclient.UpdateInvitationInput{
		ID:         id,
		Type:       toDSType(invType),
		Status:     toDSStatus(input.Status),
		AcceptedAt: normalizeTime(input.AcceptedAt),
		ExpiresAt:  normalizeTime(input.ExpiresAt),
	}
	inv, err := a.store.Update(ctx, dsInput)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, dsclient.ErrNotFound
	}
	return toDomain(inv), nil
}

func (a *dsAdapter) Revoke(ctx context.Context, id string, invType domain.InvitationType) error {
	return a.store.Revoke(ctx, id, toDSType(invType))
}

func (a *dsAdapter) Accept(ctx context.Context, token string) (*domain.Invitation, error) {
	inv, err := a.store.Accept(ctx, strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, dsclient.ErrNotFound
	}
	return toDomain(inv), nil
}

func (a *dsAdapter) Delete(ctx context.Context, id string, invType domain.InvitationType) error {
	return a.store.Delete(ctx, id, toDSType(invType))
}

func (a *dsAdapter) List(ctx context.Context, filters *domain.InvitationFilters) (*domain.PaginatedInvitations, error) {
	if filters == nil || filters.Type == nil {
		return nil, gl.Errorf("filters with type are required")
	}
	dsFilters := &dsclient.InvitationFilters{
		Type:      ptr(toDSType(*filters.Type)),
		Email:     filters.Email,
		TenantID:  filters.TenantID,
		Status:    toDSStatus(filters.Status),
		InvitedBy: filters.InvitedBy,
		Page:      filters.Page,
		Limit:     filters.Limit,
	}
	res, err := a.store.List(ctx, dsFilters)
	if err != nil {
		return nil, err
	}
	out := &domain.PaginatedInvitations{
		Data:       []*domain.Invitation{},
		Total:      res.Total,
		Page:       res.Page,
		Limit:      res.Limit,
		TotalPages: res.TotalPages,
	}
	for _, inv := range res.Data {
		out.Data = append(out.Data, toDomain(&inv))
	}
	return out, nil
}

// helpers for DS adapter -----------------------------------------------

func toDomain(inv *dsclient.Invitation) *domain.Invitation {
	if inv == nil {
		return nil
	}
	return &domain.Invitation{
		Invitation: dsclient.Invitation{
			ID:         inv.ID,
			Email:      inv.Email,
			Name:       inv.Name,
			Role:       inv.Role,
			Type:       inv.Type,
			Status:     inv.Status,
			ExpiresAt:  inv.ExpiresAt,
			AcceptedAt: inv.AcceptedAt,
			TenantID:   inv.TenantID,
			InvitedBy:  inv.InvitedBy,
			TeamID:     inv.TeamID,
			CreatedAt:  inv.CreatedAt,
			UpdatedAt:  inv.UpdatedAt,
		},
		Token: inv.Token,
	}
}

func toDSType(t domain.InvitationType) dsclient.InvitationType {
	if t == domain.TypeInternal {
		return dsclient.TypeInternal
	}
	return dsclient.TypePartner
}

func toDSStatus(status *domain.InvitationStatus) *dsclient.InvitationStatus {
	if status == nil {
		return nil
	}
	s := dsclient.InvitationStatus(*status)
	return &s
}

// func ptr[T any](v T) *T { return &v }

func normalizeTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	tt := t.UTC()
	return &tt
}

func validateInvite(inv *domain.Invitation) error {
	if inv == nil {
		return dsclient.ErrNotFound
	}
	if inv.Status != domain.StatusPending {
		return dsclient.ErrInvalidStatus
	}
	if time.Now().UTC().After(inv.ExpiresAt) {
		return dsclient.ErrExpired
	}
	return nil
}
