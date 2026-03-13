// Package dsclient defines the interface for interacting with data services.
package dsclient

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	client "github.com/kubex-ecosystem/domus/client"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	kbx "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	gl "github.com/kubex-ecosystem/logz"
)

// Type aliases para DSClient e configs
type (
	DSClient           = client.DSClient
	DSClientConfig     = client.DSClientConfig
	BackendConnections = client.BackendConnections
	BackendConnection  = client.BackendConnection
	BackendConfig      = client.BackendConfig
)

// Type aliases para Stores
type (
	StoreType             = client.StoreType
	Executor              = client.Executor
	UserStore             = client.UserStore
	InviteStore           = client.InviteStore
	CompanyStore          = client.CompanyStore
	PendingAccessStore    = client.PendingAccessStore
	ExternalMetadataStore = client.ExternalMetadataStore
)

// Type aliases para Adapter (Repository unificado Store + ORM)
type (
	AdapterFactory   = client.AdapterFactory
	RepositoryConfig = client.RepositoryConfig
)

// Type aliases genéricos (devem ser instanciados com tipo concreto)
type (
	DSRepository[T any]    = client.DSRepository[T]
	ORMRepository[T any]   = client.ORMRepository[T]
	Repository[T any]      = client.Repository[T]
	PaginatedResult[T any] = client.PaginatedResult[T]
)

// Type aliases para entidades
type (
	User                            = client.User
	CreateUserInput                 = client.CreateUserInput
	UpdateUserInput                 = client.UpdateUserInput
	UserFilters                     = client.UserFilters
	Invitation                      = client.Invitation
	InvitationType                  = client.InvitationType
	InvitationStatus                = client.InvitationStatus
	CreateInvitationInput           = client.CreateInvitationInput
	UpdateInvitationInput           = client.UpdateInvitationInput
	InvitationFilters               = client.InvitationFilters
	Company                         = client.Company
	CreateCompanyInput              = client.CreateCompanyInput
	UpdateCompanyInput              = client.UpdateCompanyInput
	CompanyFilters                  = client.CompanyFilters
	PendingAccessRequest            = client.PendingAccessRequest
	CreatePendingAccessRequestInput = client.CreatePendingAccessRequestInput
	UpdatePendingAccessRequestInput = client.UpdatePendingAccessRequestInput
	PendingAccessFilters            = client.PendingAccessFilters
	ExternalMetadataRecord          = client.ExternalMetadataRecord
	UpsertExternalMetadataInput     = client.UpsertExternalMetadataInput
	ExternalMetadataFilters         = client.ExternalMetadataFilters
)

// Invitation type constants
const (
	TypePartner  = client.TypePartner
	TypeInternal = client.TypeInternal
)

// Invitation status constants
const (
	StatusPending  = client.StatusPending
	StatusAccepted = client.StatusAccepted
	StatusRevoked  = client.StatusRevoked
	StatusExpired  = client.StatusExpired
)

var (
	// ErrNotFound sinaliza que o convite não existe.
	ErrNotFound = errors.New("invite not found")
	// ErrExpired sinaliza que o convite já expirou.
	ErrExpired = errors.New("invite expired")
	// ErrInvalidStatus indica que a operação não é permitida para o status atual.
	ErrInvalidStatus = errors.New("invite is not pending")
)

// PGExecutor expõe o executor PG do DS sem acoplar ao pacote interno.
type PGExecutor interface {
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	BeginTx(ctx context.Context) (pgx.Tx, error)
	Pool() *pgxpool.Pool
}

// Factory functions

func NewDSClient(ctx context.Context, cfg *config.MainConfig, logger *gl.LoggerZ) DSClient {
	if cfg == nil {
		gl.Errorf("nil config provided to DSClient")
		return nil
	}
	if logger == nil {
		logger = gl.GetLoggerZ("ds_client")
	}
	configPath := os.ExpandEnv(cfg.DataService.ConfigPath)
	dsConfig := NewDSClientConfig(
		cfg.ServerConfig.Name,
		cfg.ServerConfig.Files.DBConfigFile,
		nil,
	)
	return client.NewDSClient(ctx, configPath, dsConfig, logger)
}

func NewDSClientConfig(name string, filePath string, backendConfig ...*BackendConfig) *DSClientConfig {
	if len(backendConfig) == 0 {
		return client.NewDSClientConfig(name, filePath)
	}
	return client.NewDSClientConfig(name, filePath, backendConfig...)
}

func NewBackendConfig(engine, dbName, dbConfigFile string, options map[string]any) *BackendConfig {
	return client.NewBackendConfig(engine, dbName, dbConfigFile, options)
}

func NewReference(name string) *kbx.Reference {
	return kbx.NewReference(name)
}

// GetPGExecutor extrai o executor PG a partir de uma conexão retornada pelo DS.
// Evita acoplamento direto com tipos internos do domus.
func GetPGExecutor(ctx context.Context, conn *BackendConnection) (PGExecutor, error) {
	if conn == nil || conn.Driver == nil {
		return nil, gl.Errorf("nil DS connection/driver")
	}

	if conn.Driver == nil {
		return nil, gl.Errorf("nil DS driver")
	}

	exec, err := conn.Driver.Executor(ctx)
	if err != nil {
		return nil, gl.Errorf("failed to get DS executor: %v", err)
	}
	if exec == nil {
		return nil, gl.Errorf("nil DS executor")
	}

	pgExec := exec.PG()
	if pgExec == nil {
		return nil, gl.Errorf("DS connection is not a Postgres connection")
	}
	return pgExec, nil
}

// Store creation helpers

func NewUserStore(ctx context.Context, conn *BackendConnection) (UserStore, error) {
	return client.NewUserStore(ctx, conn)
}

func NewCompanyStore(ctx context.Context, conn *BackendConnection) (CompanyStore, error) {
	return client.NewCompanyStore(ctx, conn)
}

func NewInviteStore(ctx context.Context, conn *BackendConnection) (InviteStore, error) {
	return client.NewInviteStore(ctx, conn)
}

func NewPendingAccessStore(ctx context.Context, conn *BackendConnection) (PendingAccessStore, error) {
	return client.NewPendingAccessStore(ctx, conn)
}

func NewExternalMetadataStore(ctx context.Context, conn *BackendConnection) (ExternalMetadataStore, error) {
	return client.NewExternalMetadataStore(ctx, conn)
}

// Adapter config helpers

// DefaultRepositoryConfig retorna configuração padrão do adapter.
// Prefere Store com fallback automático para ORM.
func DefaultRepositoryConfig() *RepositoryConfig {
	return client.DefaultRepositoryConfig()
}

// StoreOnlyConfig retorna configuração que força uso APENAS de Store.
func StoreOnlyConfig() *RepositoryConfig {
	return client.StoreOnlyConfig()
}

// ORMOnlyConfig retorna configuração que força uso APENAS de ORM.
func ORMOnlyConfig() *RepositoryConfig {
	return client.ORMOnlyConfig()
}

// CreateAdapter cria um DSRepository[T] adapter unificado.
// Wrapper para client.CreateAdapter exportado publicamente.
func CreateAdapter[T any](
	factory *AdapterFactory,
	ctx context.Context,
	storeName string,
	ormRepoFactory func(Executor, error) ORMRepository[T],
) (*DSRepository[T], error) {
	return client.CreateAdapter[T](factory, ctx, storeName, ormRepoFactory)
}
