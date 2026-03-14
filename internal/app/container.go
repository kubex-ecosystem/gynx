// Package app contém a infraestrutura principal da aplicação, incluindo a
package app

import (
	"context"
	"crypto/rsa"
	"os"
	"path/filepath"
	"strings"

	genericapi "github.com/kubex-ecosystem/gnyx/internal/api"
	api "github.com/kubex-ecosystem/gnyx/internal/api/invite"
	ds "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore"
	companystore "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/company_store"
	ui "github.com/kubex-ecosystem/gnyx/internal/features/ui"
	kbxGet "github.com/kubex-ecosystem/kbx/get"

	userstore "github.com/kubex-ecosystem/gnyx/internal/dsclient/datastore/user_store"
	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	invitesvc "github.com/kubex-ecosystem/gnyx/internal/services/invite"
	crt "github.com/kubex-ecosystem/gnyx/internal/services/security/certificates"
	kbx "github.com/kubex-ecosystem/kbx"
	gl "github.com/kubex-ecosystem/logz"

	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/dsclient"
	"github.com/kubex-ecosystem/gnyx/internal/services/mailer"
)

// Container concentra as dependências necessárias para o gateway HTTP.
type Container struct {
	inviteSvc  api.Service                   `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	cfg        *config.Config                `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	templates  *mailer.TemplateLoader        `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	smtpSender invitesvc.MailSender          `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	db         dsclient.DSClient             `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	gormDB     *dsclient.BackendConnection   `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"` // GORM DB para fallback ORM
	factory    *dsclient.AdapterFactory      `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	stores     map[string]dsclient.StoreType `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	imapSvc    *mailer.IMAPService           `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	uiSvc      *ui.UIService                 `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`

	// Controllers genéricos CRUD usando stores do DS
	UserController    *genericapi.Controller[dsclient.User]    `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
	CompanyController *genericapi.Controller[dsclient.Company] `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
}

// NewContainer instancia a infraestrutura principal (DB, serviços de domínio, etc).
func NewContainer(ctx context.Context, cfg *config.Config) (*Container, error) {
	db, err := ds.Init(ctx, cfg)
	if err != nil {
		return nil, gl.Errorf("failed to init datastore: %v", err)
	}

	templateLoader, err := loadTemplates(cfg.ServerConfig.Files.TemplatesDir)
	if err != nil {
		return nil, gl.Errorf("failed to load templates: %v", err)
	}

	mlCfgFile := os.ExpandEnv(
		kbxGet.ValOrType(
			cfg.MailerConfigFilePath,
			kbxGet.ValOrType(
				cfg.ServerConfig.SrvConfig.Files.MailerConfigFile,
				kbxGet.EnvOr("KUBEX_GNYX_MAILER_CONFIG_PATH", kbxMod.DefaultMailConfigPath),
			),
		),
	)

	// ------------ SMTP ------------ //
	smtpMailer := buildMailer(mlCfgFile)

	// ------------ IMAP ------------ //
	imapService := buildIMAPService(mlCfgFile)

	invStore, err := ds.GetInviteStore(ctx)
	if err != nil {
		return nil, gl.Errorf("failed to create invite store: %v", err)
	}
	iAdapter, err := invitesvc.NewAdapter(invStore)
	if err != nil {
		return nil, gl.Errorf("failed to create invite adapter: %v", err)
	}
	inviteSvc, err := invitesvc.NewService(invitesvc.Config{
		Adapter:     iAdapter,
		Mailer:      buildMailer(mlCfgFile),
		Templates:   templateLoader,
		BaseURL:     cfg.Invite.BaseURL,
		SenderName:  cfg.Invite.SenderName,
		SenderEmail: cfg.Invite.SenderEmail,
		CompanyName: cfg.Invite.CompanyName,
		DefaultTTL:  cfg.Invite.DefaultTTL,
	})
	if err != nil {
		return nil, gl.Errorf("failed to create invite service: %v", err)
	}

	stores := make(map[string]dsclient.StoreType)

	usrStore, err := ds.UserStore(ctx)
	if err != nil {
		return nil, gl.Errorf("failed to create user store: %v", err)
	}
	stores["user"] = usrStore
	stores["invite"] = invStore

	uiSvc := ui.NewUIService()
	// inviteSvc.Mailer

	return &Container{
		templates:  templateLoader,
		smtpSender: smtpMailer,
		imapSvc:    imapService,
		inviteSvc:  inviteSvc,
		uiSvc:      uiSvc,

		stores: stores,

		cfg: cfg,
		db:  db,
	}, nil
}

// InviteService retorna a implementação atual do domínio de convites.
func (c *Container) InviteService() any { return c.inviteSvc }

// Config expõe a configuração para camadas que precisam do DS.
// Retorna interface{} para implementar types.IContainer e evitar import cycles.
func (c *Container) Config() any { return c.cfg }

// GetConfig retorna a configuração tipada (para uso interno no pacote app).
func (c *Container) GetConfig() *config.Config { return c.cfg }

func (c *Container) Bootstrap(ctx context.Context) error {
	gl.Debug("Bootstrapping BE...")

	// JWT certificates setup
	if c.cfg.ServerConfig.Runtime.PubCertKeyPath != "" && c.cfg.ServerConfig.Runtime.PrivKeyPath != "" {
		gl.Debug("🔐 Checking JWT certificates...")
		_, errOSPrivKey := os.Stat(c.cfg.ServerConfig.Runtime.PrivKeyPath)
		if errOSPrivKey != nil && !os.IsNotExist(errOSPrivKey) {
			return gl.Errorf("failed to access private certificate key: %v", errOSPrivKey)
		}
		_, errOSPubKey := os.Stat(c.cfg.ServerConfig.Runtime.PubCertKeyPath)
		if errOSPubKey != nil && !os.IsNotExist(errOSPubKey) {
			return gl.Errorf("failed to access public certificate key: %v", errOSPubKey)
		}

		certService := crt.NewCertServiceType(
			os.ExpandEnv(kbxGet.ValOrType(c.cfg.ServerConfig.Runtime.PrivKeyPath, kbxGet.EnvOr("KUBEX_GNYX_PRIVATE_KEY_PATH", kbxMod.DefaultGNyxKeyPath))),
			os.ExpandEnv(kbxGet.ValOrType(c.cfg.ServerConfig.Runtime.PubCertKeyPath, kbxGet.EnvOr("KUBEX_GNYX_PUBLIC_KEY_PATH", kbxMod.DefaultGNyxCertPath))),
		)

		var rsaPrivKey *rsa.PrivateKey
		if errOSPrivKey != nil || errOSPubKey != nil {
			err := os.MkdirAll(filepath.Dir(c.cfg.ServerConfig.Runtime.PrivKeyPath), 0700)
			if err != nil {
				return gl.Errorf("failed to create directory for JWT certificates: %v", err)
			}
			err = os.MkdirAll(filepath.Dir(c.cfg.ServerConfig.Runtime.PubCertKeyPath), 0700)
			if err != nil {
				return gl.Errorf("failed to create directory for JWT certificates: %v", err)
			}

			// gl.Notice("🔐 Generating new JWT certificates...")
			if _, _, _, err = certService.GenSelfCert(nil); err != nil {
				return gl.Errorf("failed to generate JWT certificates: %v", err)
			}
			gl.Debug("JWT certificates generated successfully.")
		} else {
			// gl.Debug("🔐 Loading JWT certificates...")
			var err error
			if rsaPrivKey, err = certService.DecryptPrivateKey(nil); err != nil {
				return gl.Errorf("failed to decrypt JWT private key: %v", err)
			}
		}
		gl.Noticef("🔐 JWT certificates loaded successfully. PrivKey Decrypted: %v", (rsaPrivKey != nil))
	}

	if c.db == nil {
		db, err := ds.Init(ctx, c.cfg)
		if err != nil {
			return gl.Errorf("failed to init dsclient during bootstrap: %v", err)
		}
		c.db = db
	}
	// gl.Debug("DSClient initialized successfully.")
	gormDB, err := c.initGORM(ctx)
	if err != nil {
		gl.Noticef("GORM initialization failed: %v. Stores will use DS only.", err)
		c.gormDB = nil
	} else {
		c.gormDB = gormDB
		gl.Debug("GORM initialized successfully.")
	}

	// AdapterFactory creation
	// gl.Debug("Creating AdapterFactory...")
	activeDBName, err := ds.ActiveDBName(ctx)
	if err != nil {
		return gl.Errorf("failed to resolve active DS database: %v", err)
	}
	gl.Noticef("Active DS database resolved to: %s", activeDBName)
	factory, err := c.db.NewAdapterFactory(ctx, activeDBName, c.gormDB, nil)
	if err != nil {
		return gl.Errorf("failed to create adapter factory: %v", err)
	}
	c.factory = factory
	gl.Debug("AdapterFactory created successfully.")

	// Create adapters and controllers
	if err := c.createAdaptersAndControllers(ctx); err != nil {
		return gl.Errorf("failed to create adapters and controllers: %v", err)
	}

	gl.Debug("BE bootstrapped successfully.")
	return nil
}

// initGORM inicializa GORM DB para fallback ORM.
// Temporário enquanto migra para Stores.
func (c *Container) initGORM(ctx context.Context) (*dsclient.BackendConnection, error) {
	gormDBConn, err := ds.Connection(ctx)
	if err != nil {
		return nil, err
	}
	return gormDBConn, nil
}

// createAdaptersAndControllers cria adapters e controllers genéricos.
func (c *Container) createAdaptersAndControllers(ctx context.Context) error {
	gl.Debug("Creating generic controllers...")

	// User Controller (usando UserStore do DS)
	userConn, err := ds.Connection(ctx)
	if err != nil {
		return gl.Errorf("failed to get connection: %v", err)
	}

	userStore, err := dsclient.NewUserStore(ctx, userConn)
	if err != nil {
		return gl.Errorf("failed to create user store: %v", err)
	}

	// Adiciona UserStore ao map de stores
	c.stores["user"] = userStore
	gl.Debug("UserStore created and added to stores map")

	// Cria UserStoreAdapter para normalizar assinaturas
	// UserStore.Create: (ctx, *CreateUserInput) → (*User, error)
	// Repository[User].Create: (ctx, *User) → (string, error)
	userAdapter := userstore.NewUserStoreAdapter(userStore)
	gl.Debug("UserStoreAdapter created")

	// Cria UserController genérico usando adapter
	c.UserController = genericapi.NewController(userAdapter)
	gl.Debug("UserController genérico criado com sucesso!")
	// Company Controller (usando CompanyStore do DS)
	companyStore, err := dsclient.NewCompanyStore(ctx, userConn)
	if err != nil {
		return gl.Errorf("failed to create company store: %v", err)
	}

	// Adiciona CompanyStore ao map de stores
	c.stores["company"] = companyStore
	gl.Debug("CompanyStore created and added to stores map")

	// Cria CompanyStoreAdapter para normalizar assinaturas
	companyAdapter := companystore.NewCompanyStoreAdapter(companyStore)
	gl.Debug("CompanyStoreAdapter created")
	// Cria CompanyController genérico usando adapter
	c.CompanyController = genericapi.NewController(companyAdapter)
	gl.Debug("CompanyController genérico criado com sucesso!")

	return nil
}

// GetUserController returns the user controller (interface{} for IContainer compliance).
func (c *Container) GetUserController() any {
	return c.UserController
}

// GetCompanyController returns the company controller (interface{} for IContainer compliance).
func (c *Container) GetCompanyController() any {
	return c.CompanyController
}

// GetDSClient returns the data store client (interface{} for IContainer compliance).
func (c *Container) GetDSClient(ctx context.Context) any {
	return c.db
}

// IMAPService retorna o serviço IMAP opcional.
func (c *Container) IMAPService() any {
	return c.imapSvc
}

func (c *Container) UIService() any {
	return kbxGet.ValueOrIf(!c.cfg.ServerConfig.Basic.UIDisabled, kbxGet.ValOrType(c.uiSvc, ui.NewUIService()), nil)
}

func loadTemplates(dir string) (*mailer.TemplateLoader, error) {
	if strings.TrimSpace(dir) == "" {
		return mailer.GetDefaultTemplateLoader()
	}
	loader := mailer.NewTemplateLoader(dir)
	if err := loader.LoadAll(); err != nil {
		gl.Warnf("failed to load templates from %s: %v. Falling back to defaults", dir, err)
		return mailer.GetDefaultTemplateLoader()
	}
	return loader, nil
}

func buildMailer(configPath string) invitesvc.MailSender {
	configPath = strings.TrimSpace(configPath)
	if configPath == "" {
		gl.Notice("SMTP mailer not configured: empty config path. Using noop mailer.")
		return noopMailer{}
	}
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			gl.Noticef("SMTP mailer not configured: config not found at %s. Using noop mailer.", configPath)
			return noopMailer{}
		}
		gl.Warnf("Failed to stat mailer config at %s; falling back to noop: %v", configPath, err)
		return noopMailer{}
	}

	sender, err := mailer.NewKBXSenderFromPath(configPath)
	if err != nil {
		gl.Warnf("Failed to init mailer; falling back to noop: %v", err)
		return noopMailer{}
	}
	if smtpConn := sender.Mailer.GetSMTPConnection(); smtpConn != nil {
		gl.Noticef(
			"SMTP mailer initialized: provider=%s protocol=%s host=%s port=%d from=%s",
			strings.TrimSpace(smtpConn.Provider),
			kbxGet.ValueOrIf(strings.TrimSpace(smtpConn.Protocol) != "", strings.TrimSpace(smtpConn.Protocol), "smtp"),
			strings.TrimSpace(smtpConn.Host),
			smtpConn.Port,
			strings.TrimSpace(smtpConn.From),
		)
	} else {
		gl.Warnf("SMTP mailer initialized from %s but no SMTP connection was selected", configPath)
	}
	return sender
}

func buildIMAPService(c string) *mailer.IMAPService {
	c = strings.TrimSpace(c)
	if c == "" {
		gl.Notice("IMAP service not configured: empty config path. Skipping IMAP init.")
		return nil
	}
	if _, err := os.Stat(c); err != nil {
		if os.IsNotExist(err) {
			gl.Noticef("IMAP service not configured: config not found at %s. Skipping IMAP init.", c)
			return nil
		}
		gl.Warnf("Failed to stat IMAP config at %s: %v", c, err)
		return nil
	}

	var i *mailer.IMAPService
	var err error
	mCfg, err := kbx.LoadConfigOrDefault[kbx.MailConfig](c, true)
	if err != nil {
		gl.Warnf("Failed to load MailConfig for IMAP service: %v", err)
		return nil
	}
	if mCfg == nil {
		i, err = mailer.TryIMAPFromEnv(c)
		if err != nil {
			gl.Warnf("Failed to init IMAP service: %v", err)
			return nil
		}
	} else {
		for _, conn := range mCfg.Connections {
			if strings.EqualFold(conn.Protocol, "imap") && conn.IsDefault {
				i = &mailer.IMAPService{
					IMAPConfig: &mailer.IMAPConfig{
						MailGeneralConfig:  conn.MailGeneralConfig,
						MailAuthConfig:     conn.MailAuthConfig,
						MailProtocolConfig: conn.MailProtocolConfig,
					},
				}
				break
			}
		}
	}
	if i != nil && i.IMAPConfig != nil {
		gl.Noticef(
			"IMAP service initialized: provider=%s protocol=%s host=%s port=%d mailbox=%s",
			strings.TrimSpace(i.Provider),
			kbxGet.ValueOrIf(strings.TrimSpace(i.Protocol) != "", strings.TrimSpace(i.Protocol), "imap"),
			strings.TrimSpace(i.Host),
			i.Port,
			strings.TrimSpace(i.MailBox),
		)
	} else {
		gl.Noticef("IMAP service not initialized: no default IMAP connection found in %s", c)
	}
	return i
}

type noopMailer struct{}

func (noopMailer) Send(_ *mailer.EmailMessage) error {
	gl.Debug("SMTP sender not configured. Skipping email dispatch.")
	return nil
}
