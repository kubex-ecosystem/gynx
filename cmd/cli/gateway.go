package cli

import (
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	"github.com/kubex-ecosystem/gnyx/internal/runtime/gateway"
	"github.com/spf13/cobra"

	kbx "github.com/kubex-ecosystem/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

var (
	kbxConfig *kbx.SrvConfig
	cfg       = &config.ServerConfig{
		SrvConfig: kbx.NewSrvArgs(),
	}
	Name         string
	CORSDisabled bool
	ReleaseMode  bool
)

// GatewayCmds returns the gateway command with subcommands
func GatewayCmds() *cobra.Command {
	// Main gateway command
	// var cfgFileExtensions = []string{"yaml", "yml", "jsonc", "json", "xml"}
	var initArgs = &kbxMod.InitArgs{}

	rootCmd := &cobra.Command{
		Use:   "gateway",
		Short: "GNyx Gateway - AI Provider Gateway with Repository Intelligence",
		Long: `GNyx Gateway provides a unified API for AI providers with enterprise features.

Features:
  • Multi-provider AI gateway (OpenAI, Anthropic, Gemini, Groq, etc.)
  • Repository Intelligence APIs (DORA metrics, Code Health, AI Impact)
  • Enterprise production features (rate limiting, circuit breaker, health checks)
  • Real-time streaming with Server-Sent Events (SSE)
  • BYOK (Bring Your Own Key) support
  • Tenant and user isolation`,
		Example: `  # Start gateway with default settings
  kubexbe gateway serve

  # Start with custom config and address
  kubexbe gateway serve -b '0.0.0.0' -p '5000' --config-file './config/config.example.yml'

  # Start with debug mode and CORS disable
  kubexbe gateway serve -b '0.0.0.0' -p '5000' --debug --disableCors`,
	}

	// Add subcommands
	rootCmd.AddCommand(cmdServe(initArgs))
	rootCmd.AddCommand(cmdStatus(initArgs))
	rootCmd.AddCommand(cmdAdvise(initArgs))

	return rootCmd
}

// cmdServe starts the gateway server
func cmdServe(initArgs *kbxMod.InitArgs) *cobra.Command {
	printBanner := kbxGet.EnvOrType("KUBEX_PRINT_BANNER", kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", false))
	short := "Start the gateway server (GUI Enabled)"
	long := "Start the GNyx Gateway server with enterprise features (GUI Enabled)"
	examples := []string{
		"kubexbe gateway serve -b '0.0.0.0' -p '5000' --config-file './config/config.example.yml'",
		"kubexbe gateway serve -b '0.0.0.0' -p '5000' --debug --disableCors",
	}
	// Serve subcommand
	serveCmd := &cobra.Command{
		Use:         "serve",
		Aliases:     []string{"start", "run"},
		Example:     ConcatenateExamples(examples),
		Annotations: GetDescriptions([]string{short, long}, printBanner),
		Run: func(cmd *cobra.Command, args []string) {
			gl.SetDebugMode(initArgs.Debug)

			// Check and load .env file if exists in args and file system
			if len(initArgs.EnvFile) > 0 {
				// Load specified env file, if exists
				if _, err := os.Stat(initArgs.EnvFile); err == nil {
					loadEnv(initArgs.EnvFile)
				} else {
					gl.Warnf(".env file specified but not found at %s, proceeding with existing environment variables", initArgs.EnvFile)
				}
			}
			if initArgs.Reference == nil {
				initArgs.Reference = kbxMod.NewReference(Name)
			}

			// Variable to hold server config
			var err error

			// Hydrate config file path with default if not set
			initArgs.ConfigFile = os.ExpandEnv(kbxGet.ValueOrIf(len(initArgs.ConfigFile) == 0, kbxGet.EnvOr("KUBEX_GNYX_CONFIG_PATH", initArgs.ConfigFile), initArgs.ConfigFile))

			// Load or create config file with kbx method
			kbxConfig, err = kbx.LoadConfigOrDefault[kbx.SrvConfig](initArgs.ConfigFile, true)
			if err != nil {
				gl.Errorf("Failed to load config: %v", err)
				return
			} else if kbxConfig == nil {
				gl.Noticef("No config file found, proceeding with default auto-generated config at %s", initArgs.ConfigFile)
				kbxConfigImpl := kbx.NewSrvArgs()
				kbxConfig = &kbxConfigImpl
			}

			cfg.SrvConfig = *kbxConfig

			// Validate port
			if _, err := net.LookupPort("tcp", initArgs.Port); err != nil {
				gl.Fatalf("Invalid port '%s': %v", initArgs.Port, err)
			}

			initArgs.CORSEnabled = kbxGet.BlPtr(!CORSDisabled)
			cfg.Basic.CORSEnabled = !CORSDisabled

			// Apply initArgs to server SrvConfig
			cfg.Basic.ReleaseMode = *kbxGet.ValOrType(initArgs.ReleaseMode, &ReleaseMode)
			cfg.Basic.Debug = initArgs.Debug

			cfg.Runtime.Bind = initArgs.Bind
			cfg.Runtime.Port = initArgs.Port
			// Start the gateway server

			if cfg.Basic.ReleaseMode {
				gl.Info("Starting GNyx Gateway in RELEASE mode")
			} else if cfg.Basic.Debug {
				gl.Info("Starting GNyx Gateway in DEBUG mode")
			} else {
				gl.Info("Starting GNyx Gateway in STANDARD mode")
			}

			cfg.InitArgs = initArgs

			if err := startGateway(cfg); err != nil {
				gl.Fatalf("Failed to start gateway: %v", err)
			}

			gl.Success("Gateway stopped successfully")
		},
	}

	// Add flags to serve command
	serveCmd.Flags().BoolVarP(&initArgs.Debug, "debug", "D", false, "Enable debug features (also sets log level to debug. Default: false)")
	serveCmd.Flags().StringVarP(&Name, "name", "n", "", "Set the server process application name")
	serveCmd.Flags().BoolVarP(&ReleaseMode, "release", "r", false, "Enable release mode performance optimizations (default: false)")

	serveCmd.Flags().StringVarP(&initArgs.LogFile, "log-file", "l", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_LOG_PATH", kbxMod.DefaultGNyxLogPath)), "Path to log file")
	serveCmd.Flags().StringVarP(&initArgs.EnvFile, "env-file", "e", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_ENV_PATH", kbxMod.DefaultGNyxEnvPath)), "Path to .env file for environment variables")
	serveCmd.Flags().StringVarP(&initArgs.ConfigFile, "config-file", "f", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_CONFIG_PATH", kbxMod.DefaultGNyxConfigPath)), "Path to gateway configuration file")
	serveCmd.Flags().StringVarP(&initArgs.DBConfigFile, "db-config", "d", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_DB_CONFIG_PATH", kbxMod.DefaultKubexDomusConfigPath)), "Path to database configuration file")
	serveCmd.Flags().StringVarP(&initArgs.MailConfigFile, "mail-config", "m", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_MAIL_CONFIG_PATH", kbxMod.DefaultMailConfigPath)), "Path to mail configuration file")
	serveCmd.Flags().StringVarP(&initArgs.ProvidersConfig, "providers-config", "a", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PROVIDERS_CONFIG_PATH", kbxMod.DefaultProvidersConfig)), "Path to AI providers configuration file")
	serveCmd.Flags().StringVarP(&initArgs.TemplateDir, "template-dir", "t", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_TEMPLATE_DIR", kbxMod.DefaultTemplatesDir)), "Path to templates directory")
	serveCmd.Flags().StringVarP(&initArgs.PubKeyPath, "pub-key-path", "P", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PUB_KEY_PATH", kbxMod.DefaultGNyxCAPath)), "Path to public key for JWT signing")
	serveCmd.Flags().StringVarP(&initArgs.PubCertKeyPath, "pub-cert-key-path", "C", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PUB_CERT_KEY_PATH", kbxMod.DefaultGNyxCertPath)), "Path to public certificate for TLS")
	serveCmd.Flags().StringVarP(&initArgs.PrivKeyPath, "priv-key-path", "K", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PRIV_KEY_PATH", kbxMod.DefaultGNyxKeyPath)), "Path to private key for JWT signing and TLS")

	serveCmd.Flags().BoolVarP(&CORSDisabled, "disableCors", "c", false, "Disable CORS headers in responses (CORS enabled by default)")
	serveCmd.Flags().StringVarP(&initArgs.Bind, "binding", "b", kbxGet.EnvOr("KUBEX_GNYX_BIND", "0.0.0.0"), "Server bind address (default: "+kbxGet.EnvOr("KUBEX_GNYX_BIND", "0.0.0.0")+")")
	serveCmd.Flags().StringVarP(&initArgs.Port, "port", "p", kbxGet.EnvOr("KUBEX_GNYX_PORT", "5000"), "Server port (default: "+kbxGet.EnvOr("KUBEX_GNYX_PORT", "5000")+")")

	serveCmd.Flags().StringSliceVarP(&initArgs.TrustedProxies, "trusted-proxies", "T", []string{}, "List of trusted proxies client IP resolution")

	return serveCmd
}

// cmdStatus checks the health status of the gateway server
func cmdStatus(initArgs *kbxMod.InitArgs) *cobra.Command {
	printBanner := kbxGet.EnvOrType("KUBEX_PRINT_BANNER", kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", false))
	short := "Check gateway status"
	long := "Check the health and status of the running gateway"

	statusCmd := &cobra.Command{
		Use:         "status",
		Annotations: GetDescriptions([]string{short, long}, printBanner),
		RunE:        executeStatus,
	}

	statusCmd.Flags().StringVarP(&initArgs.Bind, "binding", "b", kbxGet.EnvOr("KUBEX_GNYX_BIND", "0.0.0.0"), "Server bind address")
	statusCmd.Flags().StringVarP(&initArgs.Port, "port", "p", kbxGet.EnvOr("KUBEX_GNYX_PORT", "5000"), "Server port")

	return statusCmd
}

// cmdAdvise provides repository advice using AI providers and scorecard data
func cmdAdvise(initArgs *kbxMod.InitArgs) *cobra.Command {
	printBanner := kbxGet.EnvOrType("KUBEX_PRINT_BANNER", kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", false))
	short := "Generate repository advice using AI"
	long := "Generate repository advice using AI providers with scorecard data"

	adviseCmd := &cobra.Command{
		Use:         "advise",
		Annotations: GetDescriptions([]string{short, long}, printBanner),
		RunE:        adviseCommand,
	}

	adviseCmd.Flags().StringVarP(&initArgs.ProvidersConfig, "providers-config", "a", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PROVIDERS_CONFIG_PATH", kbxMod.DefaultProvidersConfig)), "Path to AI providers configuration file")
	adviseCmd.Flags().StringVarP(&initArgs.ScorecardPath, "scorecard", "s", "", "Path to scorecard JSON file for generating advice")

	return adviseCmd
}

// startGateway starts the gateway server with given configuration
func startGateway(config *config.ServerConfig) error {
	server, err := gateway.NewServer(config)
	if err != nil {
		gl.Fatalf("failed to create server: %v", err)
	}

	return server.Start()
}

// executeStatus checks the gateway status
func executeStatus(cmd *cobra.Command, args []string) error {
	var initArgs = &kbxMod.InitArgs{}

	// Build target URL

	initArgs.Bind = kbxGet.EnvOr("KUBEX_GNYX_BIND", "0.0.0.0")
	initArgs.Port = kbxGet.EnvOr("KUBEX_GNYX_PORT", "5000")
	initArgs.Host = net.JoinHostPort(initArgs.Bind, initArgs.Port)

	var targetAddress string
	if initArgs.Bind == "0.0.0.0" {
		targetAddress = net.JoinHostPort("localhost", initArgs.Port)
	} else {
		targetAddress = net.JoinHostPort(initArgs.Bind, initArgs.Port)
	}
	addressURL, err := url.JoinPath("http://", targetAddress, "healthz")
	if err != nil {
		return gl.Errorf("invalid gateway address: %v", err)
	}

	// Make HTTP GET request to health endpoint

	resp, err := http.Get(addressURL)
	if err != nil {
		return gl.Errorf("gateway not reachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		gl.Infof("Gateway is healthy")

		// Also check Repository Intelligence endpoints
		riResp, err := http.Get("http://" + targetAddress + "/api/v1/health")
		if err == nil && riResp.StatusCode == http.StatusOK {
			gl.Infof("Repository Intelligence APIs are available")
			riResp.Body.Close()
		} else {
			gl.Warnf("Repository Intelligence APIs not fully initialized")
		}

		return nil
	}

	return gl.Errorf("gateway unhealthy (status: %d)", resp.StatusCode)
}

// adviseCommand provides legacy advise functionality
func adviseCommand(cmd *cobra.Command, args []string) error {
	gl.Infof("🤖 Repository Advice using AI")
	gl.Infof("This command provides repository advice using scorecard data and AI providers.")
	gl.Infof("")
	gl.Infof("Usage:")
	gl.Infof("  kubexbe gateway advise --mode exec --provider openai --model gpt-4o-mini --scorecard ./scorecard.json")
	gl.Infof("")
	gl.Infof("Available modes: exec, code, ops, community")
	gl.Infof("Available providers: openai, anthropic, gemini, groq")

	// TODO: Implement full advise functionality using cmdAdvise
	return gl.Errorf("advise command not fully implemented yet")
}

// loadEnv loads environment variables from a specified .env file
func loadEnv(configPath string) {
	// Initialize environment variables, set them inside environment
	if err := godotenv.Load(configPath); err != nil {
		gl.Warnf("No .env file found at %s, proceeding with existing environment variables", configPath)
	}
	gl.Infof("Environment variables loaded from %s", configPath)
}
