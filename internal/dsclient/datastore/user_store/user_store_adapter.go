package userstore

import (
	"context"

	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
)

// UserStoreAdapter adapta UserStore para Repository[User].
//
// UserStore tem assinaturas:
// - Create(ctx, *CreateUserInput) (*User, error)
// - Update(ctx, *UpdateUserInput) (*User, error)
// - List(ctx, *UserFilters) (*PaginatedResult[User], error)
//
// Repository[User] precisa:
// - Create(ctx, *User) (string, error)
// - Update(ctx, *User) error
// - List(ctx, map[string]any) (*PaginatedResult[User], error)
//
// Este adapter faz a conversão entre as duas interfaces.
type UserStoreAdapter struct {
	store dsclient.UserStore
}

// NewUserStoreAdapter cria um adapter para UserStore.
func NewUserStoreAdapter(store dsclient.UserStore) *UserStoreAdapter {
	return &UserStoreAdapter{store: store}
}

// Create adapta Create de UserStore para Repository[User].
//
// Converte User → CreateUserInput, chama store.Create, retorna ID.
func (a *UserStoreAdapter) Create(ctx context.Context, user *dsclient.User) (string, error) {
	// Converte User para CreateUserInput
	input := &dsclient.CreateUserInput{
		Email:              user.Email,
		Name:               user.Name,
		LastName:           user.LastName,
		PasswordHash:       user.PasswordHash,
		Phone:              user.Phone,
		AvatarURL:          user.AvatarURL,
		Status:             user.Status,
		ForcePasswordReset: user.ForcePasswordReset,
	}

	// Chama UserStore.Create
	created, err := a.store.Create(ctx, input)
	if err != nil {
		return "", err
	}

	// UserStore retorna *User com ID, extrai o ID
	if created == nil {
		return "", nil
	}

	return created.ID, nil
}

// GetByID passa direto para UserStore (assinatura é compatível).
func (a *UserStoreAdapter) GetByID(ctx context.Context, id string) (*dsclient.User, error) {
	return a.store.GetByID(ctx, id)
}

// Update adapta Update de UserStore para Repository[User].
//
// Converte User → UpdateUserInput, chama store.Update, ignora retorno.
func (a *UserStoreAdapter) Update(ctx context.Context, user *dsclient.User) error {
	// Converte User para UpdateUserInput
	input := &dsclient.UpdateUserInput{
		ID:                 user.ID,
		Email:              &user.Email,
		Name:               user.Name,
		LastName:           user.LastName,
		PasswordHash:       user.PasswordHash,
		Phone:              user.Phone,
		AvatarURL:          user.AvatarURL,
		Status:             user.Status,
		ForcePasswordReset: &user.ForcePasswordReset,
	}

	// Chama UserStore.Update
	_, err := a.store.Update(ctx, input)
	return err
}

// Delete passa direto para UserStore (assinatura é compatível).
func (a *UserStoreAdapter) Delete(ctx context.Context, id string) error {
	return a.store.Delete(ctx, id)
}

// List adapta List de UserStore para Repository[User].
//
// Converte map[string]any → UserFilters, chama store.List.
func (a *UserStoreAdapter) List(ctx context.Context, filters map[string]any) (*dsclient.PaginatedResult[dsclient.User], error) {
	// Converte map para UserFilters
	var userFilters *dsclient.UserFilters
	if filters != nil {
		userFilters = &dsclient.UserFilters{}

		// Extrai filtros do map se presentes
		if email, ok := filters["email"].(string); ok {
			userFilters.Email = &email
		}
		if status, ok := filters["status"].(string); ok {
			userFilters.Status = &status
		}
		if page, ok := filters["page"].(int); ok {
			userFilters.Page = page
		}
		if limit, ok := filters["limit"].(int); ok {
			userFilters.Limit = limit
		}
	}

	// Chama UserStore.List
	return a.store.List(ctx, userFilters)
}
