// Package companystore fornece adapters para normalizar stores do DS com Repository[T].
//
// CompanyStoreAdapter adapta CompanyStore (que tem assinaturas específicas com CreateCompanyInput)
// para a interface Repository[Company] (que usa *Company diretamente).
package companystore

import (
	"context"

	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
)

// CompanyStoreAdapter adapta CompanyStore para Repository[Company].
//
// CompanyStore tem assinaturas:
// - Create(ctx, *CreateCompanyInput) (*Company, error)
// - Update(ctx, *UpdateCompanyInput) (*Company, error)
// - List(ctx, *CompanyFilters) (*PaginatedResult[Company], error)
//
// Repository[Company] precisa:
// - Create(ctx, *Company) (string, error)
// - Update(ctx, *Company) error
// - List(ctx, map[string]any) (*PaginatedResult[Company], error)
//
// Este adapter faz a conversão entre as duas interfaces.
type CompanyStoreAdapter struct {
	store dsclient.CompanyStore
}

// NewCompanyStoreAdapter cria um adapter para CompanyStore.
func NewCompanyStoreAdapter(store dsclient.CompanyStore) *CompanyStoreAdapter {
	return &CompanyStoreAdapter{store: store}
}

// Create adapta Create de CompanyStore para Repository[Company].
//
// Converte Company → CreateCompanyInput, chama store.Create, retorna ID.
func (a *CompanyStoreAdapter) Create(ctx context.Context, company *dsclient.Company) (string, error) {
	// Converte Company para CreateCompanyInput
	input := &dsclient.CreateCompanyInput{
		Name:          company.Name,
		Slug:          company.Slug,
		IsTrial:       company.IsTrial,
		IsActive:      company.IsActive,
		Domain:        company.Domain,
		Phone:         company.Phone,
		Address:       company.Address,
		PlanExpiresAt: company.PlanExpiresAt,
	}

	// Chama CompanyStore.Create
	created, err := a.store.Create(ctx, input)
	if err != nil {
		return "", err
	}

	// CompanyStore retorna *Company com ID, extrai o ID
	if created == nil {
		return "", nil
	}

	return created.ID, nil
}

// GetByID passa direto para CompanyStore (assinatura é compatível).
func (a *CompanyStoreAdapter) GetByID(ctx context.Context, id string) (*dsclient.Company, error) {
	return a.store.GetByID(ctx, id)
}

// Update adapta Update de CompanyStore para Repository[Company].
//
// Converte Company → UpdateCompanyInput, chama store.Update, ignora retorno.
func (a *CompanyStoreAdapter) Update(ctx context.Context, company *dsclient.Company) error {
	// Converte Company para UpdateCompanyInput
	input := &dsclient.UpdateCompanyInput{
		ID:            company.ID,
		Name:          &company.Name,
		Slug:          &company.Slug,
		IsTrial:       company.IsTrial,
		IsActive:      company.IsActive,
		Domain:        company.Domain,
		Phone:         company.Phone,
		Address:       company.Address,
		PlanExpiresAt: company.PlanExpiresAt,
	}

	// Chama CompanyStore.Update
	_, err := a.store.Update(ctx, input)
	return err
}

// Delete passa direto para CompanyStore (assinatura é compatível).
func (a *CompanyStoreAdapter) Delete(ctx context.Context, id string) error {
	return a.store.Delete(ctx, id)
}

// List adapta List de CompanyStore para Repository[Company].
//
// Converte map[string]any → CompanyFilters, chama store.List.
func (a *CompanyStoreAdapter) List(ctx context.Context, filters map[string]any) (*dsclient.PaginatedResult[dsclient.Company], error) {
	// Converte map para CompanyFilters
	var companyFilters *dsclient.CompanyFilters
	if filters != nil {
		companyFilters = &dsclient.CompanyFilters{}

		// Extrai filtros do map se presentes
		if name, ok := filters["name"].(string); ok {
			companyFilters.Name = &name
		}
		if slug, ok := filters["slug"].(string); ok {
			companyFilters.Slug = &slug
		}
		if isActive, ok := filters["is_active"].(bool); ok {
			companyFilters.IsActive = &isActive
		}
		if page, ok := filters["page"].(int); ok {
			companyFilters.Page = page
		}
		if limit, ok := filters["limit"].(int); ok {
			companyFilters.Limit = limit
		}
	}

	// Chama CompanyStore.List
	return a.store.List(ctx, companyFilters)
}
