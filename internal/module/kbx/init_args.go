// Package kbx provides utilities for working with initialization arguments.
package kbx

import (
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/kubex-ecosystem/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
)

var (
	srvCfg *kbx.SrvConfig
	Args   *InitArgs
)

// Reference is the internal struct that holds the server startup unique identifier with name.
type Reference struct {
	// refID is the unique identifier for this context.
	ID uuid.UUID
	// refName is the name of the context.
	Name string
}

func (r *Reference) GetReference() *Reference {
	return r
}

func (r *Reference) GetID() uuid.UUID {
	return r.ID
}

func (r *Reference) GetName() string {
	return r.Name
}

func (r *Reference) GetType() string {
	return "kbx_init_args"
}

func NewReference(name string) *Reference {
	return &Reference{
		ID:   uuid.New(),
		Name: name,
	}
}

type InitArgs struct {
	*Reference     `yaml:"-" json:"-" mapstructure:"-"`
	*kbx.SrvConfig `yaml:"-" json:"-" mapstructure:"-"`

	// // Basic options

	// Debug          bool     `yaml:"debug" json:"debug" mapstructure:"debug"`
	// ReleaseMode    *bool    `yaml:"release_mode" json:"release_mode" mapstructure:"release_mode"`
	// IsConfidential bool     `yaml:"is_confidential" json:"is_confidential" mapstructure:"is_confidential"`
	// CORSEnabled    *bool    `yaml:"enable_cors" json:"enable_cors" mapstructure:"enable_cors"`
	// TrustedProxies []string `yaml:"trusted_proxies" json:"trusted_proxies" mapstructure:"trusted_proxies"`
	// UIEnabled      bool     `yaml:"ui_enabled" json:"ui_enabled" mapstructure:"ui_enabled"`

	// // Paths and files

	// Cwd              string `yaml:"cwd,omitempty" json:"cwd,omitempty" mapstructure:"cwd,omitempty"`
	// LogFile          string `yaml:"log_file,omitempty" json:"log_file,omitempty" mapstructure:"log_file,omitempty"`
	// EnvFile          string `yaml:"env_file,omitempty" json:"env_file,omitempty" mapstructure:"env_file,omitempty"`
	// ConfigFile       string `yaml:"config_file,omitempty" json:"config_file,omitempty" mapstructure:"config_file,omitempty"`
	// DBConfigFile     string `yaml:"db_config_file,omitempty" json:"db_config_file,omitempty" mapstructure:"db_config_file,omitempty"`
	// MailerConfigFile string `yaml:"mail_config_file,omitempty" json:"mail_config_file,omitempty" mapstructure:"mail_config_file,omitempty"`
	// ProvidersConfig  string `yaml:"providers_config,omitempty" json:"providers_config,omitempty" mapstructure:"providers_config,omitempty"`
	// ScorecardPath    string `yaml:"scorecard_path,omitempty" json:"scorecard_path,omitempty" mapstructure:"scorecard_path,omitempty"`
	// TemplatesDir     string `yaml:"template_dir,omitempty" json:"template_dir,omitempty" mapstructure:"template_dir,omitempty"`

	// // Runtime options

	// Host            string        `yaml:"host,omitempty" json:"host,omitempty" mapstructure:"host,omitempty"`
	// Port            string        `yaml:"port,omitempty" json:"port,omitempty" mapstructure:"port,omitempty"`
	// Bind            string        `yaml:"bind,omitempty" json:"bind,omitempty" mapstructure:"bind,omitempty"`
	// PubCertKeyPath  string        `yaml:"pub_cert_key_path,omitempty" json:"pub_cert_key_path,omitempty" mapstructure:"pub_cert_key_path,omitempty"`
	// PubKeyPath      string        `yaml:"pub_key_path,omitempty" json:"pub_key_path,omitempty" mapstructure:"pub_key_path,omitempty"`
	// PrivKeyPath     string        `yaml:"priv_key_path,omitempty" json:"priv_key_path,omitempty" mapstructure:"priv_key_path,omitempty"`
	// AccessTokenTTL  time.Duration `yaml:"access_token_ttl,omitempty" json:"access_token_ttl,omitempty" mapstructure:"access_token_ttl,omitempty"`
	// RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl,omitempty" json:"refresh_token_ttl,omitempty" mapstructure:"refresh_token_ttl,omitempty"`
	// Issuer          string        `yaml:"issuer,omitempty" json:"issuer,omitempty" mapstructure:"issuer,omitempty"`

	// // Advanced options

	// Context    string            `yaml:"context,omitempty" json:"context,omitempty" mapstructure:"context,omitempty"`
	// Command    string            `yaml:"command,omitempty" json:"command,omitempty" mapstructure:"command,omitempty"`
	// Subcommand string            `yaml:"subcommand,omitempty" json:"subcommand,omitempty" mapstructure:"subcommand,omitempty"`
	// Args       string            `yaml:"args,omitempty" json:"args,omitempty" mapstructure:"args,omitempty"`
	// EnvVars    map[string]string `yaml:"env_vars,omitempty" json:"env_vars,omitempty" mapstructure:"env_vars,omitempty"`

	// // Flags

	// FailFast  bool `yaml:"fail_fast,omitempty" json:"fail_fast,omitempty" mapstructure:"fail_fast,omitempty"`
	// Verbose   bool `yaml:"verbose,omitempty" json:"verbose,omitempty" mapstructure:"verbose,omitempty"`
	// BatchMode bool `yaml:"batch_mode,omitempty" json:"batch_mode,omitempty" mapstructure:"batch_mode,omitempty"`
	// NoColor   bool `yaml:"no_color,omitempty" json:"no_color,omitempty" mapstructure:"no_color,omitempty"`
	// TraceMode bool `yaml:"trace_mode,omitempty" json:"trace_mode,omitempty" mapstructure:"trace_mode,omitempty"`
	// RootMode  bool `yaml:"root_mode,omitempty" json:"root_mode,omitempty" mapstructure:"root_mode,omitempty"`

	// // Performance options

	// MaxProcs  int    `yaml:"max_procs,omitempty" json:"max_procs,omitempty" mapstructure:"max_procs,omitempty"`
	// TimeoutMS int    `yaml:"timeout_ms,omitempty" json:"timeout_ms,omitempty" mapstructure:"timeout_ms,omitempty"`
	// Hash      string `yaml:"hash,omitempty" json:"hash,omitempty" mapstructure:"hash,omitempty"`
}

func NewArgs(name string) *InitArgs {
	return kbxGet.ValOrType(Args, &InitArgs{
		Reference: NewReference(name).GetReference(),
		SrvConfig: kbx.NewSrvCfg(),
	})
}

func InitArgsDefaults() {
	Args = kbxGet.ValOrType(Args, NewArgs(DefaultName))

	Args.Runtime = kbxGet.ValOrType(Args.Runtime, kbx.NewSrvArgs().Runtime)
	Args.Basic = kbxGet.ValOrType(Args.Basic, kbx.NewSrvArgs().Basic)
	Args.Files = kbxGet.ValOrType(Args.Files, kbx.NewSrvArgs().Files)
	Args.Performance = kbxGet.ValOrType(Args.Performance, kbx.NewSrvArgs().Performance)
	Args.Flags = kbxGet.ValOrType(Args.Flags, kbx.NewSrvArgs().Flags)

	Args.Runtime.Port = kbxGet.ValOrType(Args.Runtime.Port, kbxGet.EnvOr("KUBEX_GNYX_PORT", DefaultServerPort))
	Args.Runtime.Bind = kbxGet.ValOrType(Args.Runtime.Bind, kbxGet.EnvOr("KUBEX_GNYX_BIND", DefaultServerBind))
	Args.Runtime.Host = kbxGet.ValOrType(Args.Runtime.Host, kbxGet.EnvOr("KUBEX_GNYX_HOST", DefaultServerHost))

	Args.Basic.Debug = kbxGet.ValOrType(Args.Basic.Debug, kbxGet.EnvOrType("KUBEX_GNYX_DEBUG", false))
	Args.Basic.CORSEnabled = kbxGet.ValOrType(Args.Basic.CORSEnabled, kbxGet.EnvOrType("KUBEX_GNYX_CORS_ENABLED", false))
	Args.Basic.ReleaseMode = kbxGet.ValOrType(Args.Basic.ReleaseMode, kbxGet.EnvOrType("KUBEX_GNYX_RELEASE_MODE", false))
	Args.Basic.IsConfidential = kbxGet.ValOrType(Args.Basic.IsConfidential, kbxGet.EnvOrType("KUBEX_GNYX_IS_CONFIDENTIAL", false))

	// Args.Basic.UIEnabled = kbxGet.ValOrType(Args.Basic.UIEnabled, kbxGet.EnvOrType("KUBEX_GNYX_UI_ENABLED", false))
	Args.Runtime.AccessTokenTTL = kbxGet.ValOrType(Args.Runtime.AccessTokenTTL, kbxGet.EnvOrType[time.Duration]("KUBEX_GNYX_ACCESS_TOKEN_TTL", DefaultAccessTokenTTL))
	Args.Runtime.RefreshTokenTTL = kbxGet.ValOrType(Args.Runtime.RefreshTokenTTL, kbxGet.EnvOrType[time.Duration]("KUBEX_GNYX_REFRESH_TOKEN_TTL", DefaultRefreshTokenTTL))
	Args.Runtime.PubCertKeyPath = os.ExpandEnv(kbxGet.ValOrType(Args.Runtime.PubCertKeyPath, kbxGet.EnvOr("KUBEX_GNYX_KEY_PATH", DefaultKeyPath)))
	Args.Runtime.PubKeyPath = os.ExpandEnv(kbxGet.ValOrType(Args.Runtime.PubKeyPath, kbxGet.EnvOr("KUBEX_GNYX_CERT_PATH", DefaultCertPath)))
	Args.Runtime.Issuer = kbxGet.ValOrType(Args.Runtime.Issuer, kbxGet.EnvOr("KUBEX_GNYX_ISSUER", DefaultIssuer))

	Args.Files.ConfigFile = os.ExpandEnv(kbxGet.ValOrType(Args.Files.ConfigFile, kbxGet.EnvOr("KUBEX_GNYX_CONFIG_FILE_PATH", DefaultConfigPath)))
	Args.Files.DBConfigFile = os.ExpandEnv(kbxGet.ValOrType(Args.Files.DBConfigFile, kbxGet.EnvOr("KUBEX_GNYX_DOMUS_CONFIG_PATH", DefaultDomusConfigPath)))
	Args.Files.EnvFile = os.ExpandEnv(kbxGet.ValOrType(Args.Files.EnvFile, kbxGet.EnvOr("KUBEX_GNYX_DOMUS_CONFIG_PATH", DefaultDomusConfigPath)))
	Args.Files.LogFile = os.ExpandEnv(kbxGet.ValOrType(Args.Files.LogFile, kbxGet.EnvOr("KUBEX_GNYX_LOG_FILE_PATH", DefaultLogPath)))
	Args.Files.TemplatesDir = os.ExpandEnv(kbxGet.ValOrType(Args.Files.TemplatesDir, kbxGet.EnvOr("KUBEX_GNYX_TEMPLATES_DIR", DefaultTemplatesDir)))
	// Args.Files.ScorecardPath = os.ExpandEnv(kbxGet.ValOrType(Args.Files.ScorecardPath, kbxGet.EnvOr("KUBEX_GNYX_SCORECARD_PATH", DefaultScorecardPath)))

	Args.Performance.TimeoutMS = kbxGet.ValOrType(Args.Performance.TimeoutMS, kbxGet.EnvOrType("KUBEX_GNYX_TIMEOUT_MS", DefaultTimeoutMS))
	Args.Performance.MaxProcs = kbxGet.ValOrType(Args.Performance.MaxProcs, kbxGet.EnvOrType("KUBEX_GNYX_MAX_PROCS", DefaultMaxProcs))

	Args.Flags.NoColor = kbxGet.ValOrType(Args.Flags.NoColor, kbxGet.EnvOrType("KUBEX_GNYX_NO_COLOR", false))
	Args.Flags.TraceMode = kbxGet.ValOrType(Args.Flags.TraceMode, kbxGet.EnvOrType("KUBEX_GNYX_TRACE_MODE", false))
	Args.Flags.RootMode = kbxGet.ValOrType(Args.Flags.RootMode, kbxGet.EnvOrType("KUBEX_GNYX_ROOT_MODE", false))
	Args.Flags.BatchMode = kbxGet.ValOrType(Args.Flags.BatchMode, kbxGet.EnvOrType("KUBEX_GNYX_BATCH_MODE", false))
	Args.Flags.FailFast = kbxGet.ValOrType(Args.Flags.FailFast, kbxGet.EnvOrType("KUBEX_GNYX_FAIL_FAST", false))
	Args.Flags.Verbose = kbxGet.ValOrType(Args.Flags.Verbose, kbxGet.EnvOrType("KUBEX_GNYX_VERBOSE", false))
}
