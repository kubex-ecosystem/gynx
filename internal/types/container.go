// Package types provides common interfaces to avoid import cycles.
package types

import "context"

// IContainer is the interface that app.Container implements.
// This interface is used by routes and other packages to avoid import cycles.
//
// The app.Container implements this interface and provides dependency injection
// for all application components.
type IContainer interface {
	// Config returns the application configuration (interface{} to avoid import cycle)
	Config() interface{}

	// Bootstrap initializes the application container
	Bootstrap(ctx context.Context) error

	// GetDSClient returns the data store client
	GetDSClient(ctx context.Context) interface{}

	// Controller getters (generic controllers for CRUD operations)
	GetUserController() interface{}
	GetCompanyController() interface{}

	// Service getters
	InviteService() interface{}
	IMAPService() interface{}
}
