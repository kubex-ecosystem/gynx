package invitestore

import (
	"context"
	"strings"
	"time"

	domain "github.com/kubex-ecosystem/gnyx/internal/domain/invites"
	ds "github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore"
	gl "github.com/kubex-ecosystem/logz"
)

// InviteBridge adapta o InviteStore do DS para o domínio atual.
type InviteBridge struct {
	store ds.InviteStore
}

func NewInviteBridge(ctx context.Context) (*InviteBridge, error) {
	store, err := datastore.GetInviteStore(ctx)
	if err != nil {
		return nil, err
	}
	return &InviteBridge{store: store}, nil
}

func (b *InviteBridge) GetByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	inv, err := b.store.GetByToken(ctx, strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ds.ErrNotFound
	}
	domainInv := toDomain(inv)
	if err := validateInvite(domainInv); err != nil {
		return nil, err
	}
	return domainInv, nil
}

func (b *InviteBridge) GetByID(ctx context.Context, id string, invType domain.InvitationType) (*domain.Invitation, error) {
	dsType := toDSType(invType)
	inv, err := b.store.GetByID(ctx, id, dsType)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ds.ErrNotFound
	}
	return toDomain(inv), nil
}

func (b *InviteBridge) CreatePartner(ctx context.Context, input *domain.CreatePartnerInvitationInput) (*domain.Invitation, error) {
	if input == nil {
		return nil, gl.Errorf("input is required")
	}
	dsInput := &ds.CreateInvitationInput{
		Type:      ds.TypePartner,
		Name:      safeString(input.PartnerName, input.PartnerEmail),
		Email:     strings.ToLower(strings.TrimSpace(input.PartnerEmail)),
		Role:      input.Role,
		Token:     input.Token,
		TenantID:  input.TenantID,
		TeamID:    nil,
		InvitedBy: input.InvitedBy,
		ExpiresAt: normalizeTime(input.ExpiresAt),
	}
	inv, err := b.store.Create(ctx, dsInput)
	if err != nil {
		return nil, err
	}
	return toDomain(inv), nil
}

func (b *InviteBridge) CreateInternal(ctx context.Context, input *domain.CreateInternalInvitationInput) (*domain.Invitation, error) {
	if input == nil {
		return nil, gl.Errorf("input is required")
	}
	dsInput := &ds.CreateInvitationInput{
		Type:      ds.TypeInternal,
		Name:      safeString(input.InviteeName, input.InviteeEmail),
		Email:     strings.ToLower(strings.TrimSpace(input.InviteeEmail)),
		Role:      input.Role,
		Token:     input.Token,
		TenantID:  input.TenantID,
		TeamID:    input.TeamID,
		InvitedBy: input.InvitedBy,
		ExpiresAt: normalizeTime(input.ExpiresAt),
	}
	inv, err := b.store.Create(ctx, dsInput)
	if err != nil {
		return nil, err
	}
	return toDomain(inv), nil
}

func (b *InviteBridge) Update(ctx context.Context, id string, invType domain.InvitationType, input *domain.UpdateInvitationInput) (*domain.Invitation, error) {
	if input == nil {
		return nil, gl.Errorf("update input is required")
	}
	dsInput := &ds.UpdateInvitationInput{
		ID:         id,
		Type:       toDSType(invType),
		Status:     toDSStatus(input.Status),
		AcceptedAt: normalizeTime(input.AcceptedAt),
		ExpiresAt:  normalizeTime(input.ExpiresAt),
	}
	inv, err := b.store.Update(ctx, dsInput)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ds.ErrNotFound
	}
	return toDomain(inv), nil
}

func (b *InviteBridge) Revoke(ctx context.Context, id string, invType domain.InvitationType) error {
	return b.store.Revoke(ctx, id, toDSType(invType))
}

func (b *InviteBridge) Accept(ctx context.Context, token string) (*domain.Invitation, error) {
	inv, err := b.store.Accept(ctx, strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ds.ErrNotFound
	}
	return toDomain(inv), nil
}

func (b *InviteBridge) Delete(ctx context.Context, id string, invType domain.InvitationType) error {
	return b.store.Delete(ctx, id, toDSType(invType))
}

func (b *InviteBridge) List(ctx context.Context, filters *domain.InvitationFilters) (*domain.PaginatedInvitations, error) {
	if filters == nil || filters.Type == nil {
		return nil, gl.Errorf("filters with type are required")
	}

	dsFilters := &ds.InvitationFilters{
		Type:      ptr(toDSType(*filters.Type)),
		Email:     filters.Email,
		TenantID:  filters.TenantID,
		Status:    toDSStatus(filters.Status),
		InvitedBy: filters.InvitedBy,
		Page:      filters.Page,
		Limit:     filters.Limit,
	}

	res, err := b.store.List(ctx, dsFilters)
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

func toDomain(inv *ds.Invitation) *domain.Invitation {
	if inv == nil {
		return nil
	}
	return &domain.Invitation{
		Invitation: ds.Invitation{
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

func toDSType(t domain.InvitationType) ds.InvitationType {
	if t == domain.TypeInternal {
		return ds.TypeInternal
	}
	return ds.TypePartner
}

func toDSStatus(status *domain.InvitationStatus) *ds.InvitationStatus {
	if status == nil {
		return nil
	}
	s := ds.InvitationStatus(*status)
	return &s
}

func normalizeTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	tt := t.UTC()
	return &tt
}

func safeString(primary *string, fallback string) string {
	if primary != nil && strings.TrimSpace(*primary) != "" {
		return strings.TrimSpace(*primary)
	}
	return strings.TrimSpace(fallback)
}

func ptr[T any](v T) *T { return &v }

func validateInvite(inv *domain.Invitation) error {
	if inv == nil {
		return ds.ErrNotFound
	}
	if inv.Status != domain.StatusPending {
		return ds.ErrInvalidStatus
	}
	if time.Now().UTC().After(inv.ExpiresAt) {
		return ds.ErrExpired
	}
	return nil
}
