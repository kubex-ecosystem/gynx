package cli

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/runtime/gateway"
	"github.com/spf13/cobra"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	vs "github.com/kubex-ecosystem/gnyx/internal/module/version"
	kbx "github.com/kubex-ecosystem/kbx"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

var (
	Name         string
	CORSDisabled bool
	ReleaseMode  bool
)

// GatewayCmds returns the gateway command with subcommands
func GatewayCmds() *cobra.Command {
	kbxMod.InitArgsDefaults()

	printBanner := kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", false))
	short := "GNyx Gateway - AI Provider Gateway with Repository Intelligence"
	long := `GNyx Gateway provides a unified API for AI providers with enterprise features.

Features:
  • Multi-provider AI gateway (OpenAI, Anthropic, Gemini, Groq, etc.)
  • Repository Intelligence APIs (DORA metrics, Code Health, AI Impact)
  • Enterprise production features (rate limiting, circuit breaker, health checks)
  • Real-time streaming with Server-Sent Events (SSE)
  • BYOK (Bring Your Own Key) support
  • Tenant and user isolation`

	rootCmd := &cobra.Command{
		Use:         "gateway",
		Aliases:     []string{"gw", "srv", "server"},
		Annotations: GetDescriptions([]string{short, long}, printBanner),
		Example: `# Start gateway with default settings
  gnyx gateway up

  # Start with custom config and address
  gnyx gateway up -b '0.0.0.0' -p '5000' --config-file './config/config.example.yml'

  # Start with debug mode and CORS disable
  gnyx gateway up -b '0.0.0.0' -p '5000' --debug --disableCors
	`,
	}

	// Add subcommands
	rootCmd.AddCommand(cmdUp())
	rootCmd.AddCommand(cmdDown())
	rootCmd.AddCommand(cmdStatus())
	rootCmd.AddCommand(cmdAdvise())

	return rootCmd
}

// cmdUp starts the gateway server
func cmdUp() *cobra.Command {
	initArgs := kbxMod.Args

	printBanner := kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", false))
	short := "Start the gateway server (GUI Enabled)"
	long := "Start the GNyx Gateway server with enterprise features (GUI Enabled)"
	examples := []string{
		"gnyx gateway up -b '0.0.0.0' -p '5000' --config-file './config/config.example.yml'",
		"gnyx gateway up -b '0.0.0.0' -p '5000' --debug --disableCors",
	}
	// Up subcommand
	upCmd := &cobra.Command{
		Use:         "up",
		Aliases:     []string{"run", "start", "serve"},
		Example:     ConcatenateExamples(examples),
		Annotations: GetDescriptions([]string{short, long}, printBanner),
		Version:     kbxGet.ValOrType(vs.GetVersion(), "unknown"),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			// Initialize default values for args
			kbxMod.InitArgsDefaults()

			// Set debug mode
			gl.SetDebugMode(initArgs.Basic.Debug)

			// Check and load .env file if exists in args and file system
			initArgs.Files.EnvFile = os.ExpandEnv(kbxGet.ValOrType(initArgs.Files.EnvFile, kbxGet.EnvOr("KUBEX_GNYX_ENV_PATH", kbxMod.DefaultEnvPath)))

			// Load specified env file, if exists
			if _, err := os.Stat(initArgs.Files.EnvFile); err == nil {
				loadEnv(initArgs.Files.EnvFile)
			} else {
				gl.Warnf(".env file specified but not found at %s, proceeding with existing environment variables", initArgs.Files.EnvFile)
				gl.Debugf("Env file load error: %v", err)
			}

			// Set config file path
			initArgs.Files.ConfigFile = os.ExpandEnv(kbxGet.ValOrType(initArgs.Files.ConfigFile, kbxGet.EnvOr("KUBEX_GNYX_CONFIG_PATH", kbxMod.DefaultConfigPath)))

			// Load or create config file with kbx method
			kbxConfig, err := kbx.LoadConfigOrDefault[kbx.SrvConfig](initArgs.Files.ConfigFile, true)
			if err != nil {
				gl.Errorf("Failed to load config: %v", err)
				return
			}
			cfg := config.NewServerConfig()
			cfg.SrvConfig = kbxGet.ValOrType(kbxConfig, initArgs.SrvConfig)

			// Validate port
			if _, err := net.LookupPort("tcp", cfg.Runtime.Port); err != nil {
				gl.Fatalf("Invalid port '%s': %v", cfg.Runtime.Port, err)
			}
			// Apply initArgs to server SrvConfig
			cfg.Basic.CORSEnabled = !CORSDisabled
			cfg.Basic.ReleaseMode = kbxGet.ValOrType(initArgs.Basic.ReleaseMode, ReleaseMode)
			cfg.Basic.Debug = kbxGet.ValOrType(cfg.Basic.Debug, initArgs.Basic.Debug)
			cfg.Runtime.Bind = kbxGet.ValOrType(cfg.Runtime.Bind, initArgs.Runtime.Bind)
			cfg.Runtime.Port = kbxGet.ValOrType(cfg.Runtime.Port, initArgs.Runtime.Port)
			cfg.Files.ConfigFile = kbxGet.ValOrType(cfg.Files.ConfigFile, initArgs.Files.ConfigFile)
			cfg.Files.DBConfigFile = kbxGet.ValOrType(cfg.Files.DBConfigFile, initArgs.Files.DBConfigFile)
			cfg.Files.MailerConfigFile = kbxGet.ValOrType(cfg.Files.MailerConfigFile, initArgs.Files.MailerConfigFile)
			cfg.Files.ProvidersConfig = kbxGet.ValOrType(cfg.Files.ProvidersConfig, initArgs.Files.ProvidersConfig)
			cfg.Files.TemplatesDir = kbxGet.ValOrType(cfg.Files.TemplatesDir, initArgs.Files.TemplatesDir)
			cfg.Runtime.PubKeyPath = kbxGet.ValOrType(cfg.Runtime.PubKeyPath, initArgs.Runtime.PubKeyPath)
			cfg.Runtime.PubCertKeyPath = kbxGet.ValOrType(cfg.Runtime.PubCertKeyPath, initArgs.Runtime.PubCertKeyPath)
			cfg.Runtime.PrivKeyPath = kbxGet.ValOrType(cfg.Runtime.PrivKeyPath, initArgs.Runtime.PrivKeyPath)
			cfg.InitArgs = initArgs

			// Start the gateway server
			if cfg.Basic.ReleaseMode {
				gl.Info("Starting GNyx Gateway in RELEASE mode")
			} else if cfg.Basic.Debug {
				gl.Info("Starting GNyx Gateway in DEBUG mode")
			} else {
				gl.Info("Starting GNyx Gateway in STANDARD mode")
			}

			if err := startGateway(cfg); err != nil {
				gl.Fatalf("Failed to start gateway: %v", err)
			}

			gl.Success("GNyx Gateway stopped successfully")
		},
	}

	// Add flags to serve command
	upCmd.Flags().StringVarP(&Name, "name", "n", "", "Set the server process application name")
	upCmd.Flags().BoolVarP(&initArgs.Basic.Debug, "debug", "D", false, "Enable debug features (also sets log level to debug. Default: false)")
	upCmd.Flags().BoolVarP(&ReleaseMode, "production", "r", false, "Enable release mode performance optimizations (default: false)")

	upCmd.Flags().StringVarP(&initArgs.Files.LogFile, "log-file", "l", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_LOG_PATH", kbxMod.DefaultLogPath)), "Path to log file")
	upCmd.Flags().StringVarP(&initArgs.Files.EnvFile, "env-file", "e", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_ENV_PATH", kbxMod.DefaultEnvPath)), "Path to .env file for environment variables")
	upCmd.Flags().StringVarP(&initArgs.Files.ConfigFile, "config-file", "f", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_CONFIG_PATH", kbxMod.DefaultConfigPath)), "Path to gateway configuration file")
	upCmd.Flags().StringVarP(&initArgs.Files.DBConfigFile, "db-config", "d", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_DB_CONFIG_PATH", kbxMod.DefaultDomusConfigPath)), "Path to database configuration file")
	upCmd.Flags().StringVarP(&initArgs.Files.MailerConfigFile, "mail-config", "m", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_MAIL_CONFIG_PATH", kbxMod.DefaultMailConfigPath)), "Path to mail configuration file")
	upCmd.Flags().StringVarP(&initArgs.Files.ProvidersConfig, "providers-config", "a", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PROVIDERS_CONFIG_PATH", kbxMod.DefaultProvidersConfigPath)), "Path to AI providers configuration file")
	upCmd.Flags().StringVarP(&initArgs.Files.TemplatesDir, "template-dir", "t", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_TEMPLATE_DIR", kbxMod.DefaultTemplatesDir)), "Path to templates directory")
	upCmd.Flags().StringVarP(&initArgs.Runtime.PubKeyPath, "pub-key-path", "P", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PUB_KEY_PATH", kbxMod.DefaultCAPath)), "Path to public key for JWT signing")
	upCmd.Flags().StringVarP(&initArgs.Runtime.PubCertKeyPath, "pub-cert-key-path", "C", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PUB_CERT_KEY_PATH", kbxMod.DefaultCertPath)), "Path to public certificate for TLS")
	upCmd.Flags().StringVarP(&initArgs.Runtime.PrivKeyPath, "priv-key-path", "K", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PRIV_KEY_PATH", kbxMod.DefaultKeyPath)), "Path to private key for JWT signing and TLS")

	upCmd.Flags().BoolVarP(&CORSDisabled, "disableCors", "c", false, "Disable CORS headers in responses (CORS enabled by default)")
	upCmd.Flags().StringVarP(&initArgs.Runtime.Bind, "binding", "b", kbxGet.EnvOr("KUBEX_GNYX_BIND", "0.0.0.0"), "Server bind address (default: "+kbxGet.EnvOr("KUBEX_GNYX_BIND", "0.0.0.0")+")")
	upCmd.Flags().StringVarP(&initArgs.Runtime.Port, "port", "p", kbxGet.EnvOr("KUBEX_GNYX_PORT", "5000"), "Server port (default: "+kbxGet.EnvOr("KUBEX_GNYX_PORT", "5000")+")")

	upCmd.Flags().StringSliceVarP(&initArgs.Basic.TrustedProxies, "trusted-proxies", "T", []string{}, "List of trusted proxies client IP resolution")

	return upCmd
}

// cmdDown stops the gateway server
func cmdDown() *cobra.Command {
	initArgs := kbxMod.Args

	printBanner := kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", false))
	short := "Stop the gateway server"
	long := "Stop the GNyx Gateway server gracefully"
	examples := []string{
		"gnyx gateway down",
	}
	downCmd := &cobra.Command{
		Use:         "down",
		Aliases:     []string{"stop"},
		Example:     ConcatenateExamples(examples),
		Annotations: GetDescriptions([]string{short, long}, printBanner),
		Run: func(cmd *cobra.Command, args []string) {
			gl.SetDebugMode(initArgs.Basic.Debug)
			gl.Info("Stopping GNyx Gateway...")

			if err := stopGateway(); err != nil {
				gl.Errorf("Failed to stop gateway: %v", err)
			} else {
				gl.Success("GNyx Gateway stopped successfully")
			}
		},
	}

	return downCmd
}

// cmdStatus checks the health status of the gateway server
func cmdStatus() *cobra.Command {
	initArgs := kbxMod.Args

	printBanner := kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", false))
	short := "Check gateway status"
	long := "Check the health and status of the running gateway"

	statusCmd := &cobra.Command{
		Use:         "status",
		Annotations: GetDescriptions([]string{short, long}, printBanner),
		RunE:        executeStatus,
	}

	statusCmd.Flags().StringVarP(&initArgs.Runtime.Bind, "binding", "b", kbxGet.EnvOr("KUBEX_GNYX_BIND", "0.0.0.0"), "Server bind address")
	statusCmd.Flags().StringVarP(&initArgs.Runtime.Port, "port", "p", kbxGet.EnvOr("KUBEX_GNYX_PORT", "5000"), "Server port")

	return statusCmd
}

// cmdAdvise provides repository advice using AI providers and scorecard data
func cmdAdvise() *cobra.Command {
	initArgs := kbxMod.Args

	printBanner := kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", kbxGet.EnvOrType("KUBEX_GNYX_PRINT_BANNER", false))
	short := "Generate repository advice using AI"
	long := "Generate repository advice using AI providers with scorecard data"

	adviseCmd := &cobra.Command{
		Use:         "advise",
		Annotations: GetDescriptions([]string{short, long}, printBanner),
		RunE:        adviseCommand,
	}

	adviseCmd.Flags().StringVarP(&initArgs.Files.ProvidersConfig, "providers-config", "a", os.ExpandEnv(kbxGet.EnvOr("KUBEX_GNYX_PROVIDERS_CONFIG_PATH", kbxMod.DefaultProvidersConfigPath)), "Path to AI providers configuration file")
	// adviseCmd.Flags().StringVarP(&initArgs.Files.ScorecardPath, "scorecard", "s", "", "Path to scorecard JSON file for generating advice")

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

// stopGateway stops the gateway server gracefully
func stopGateway() error {
	initArgs := kbxMod.Args
	// For simplicity, this example assumes the gateway can be stopped by sending a shutdown signal to the server process.
	// In a real implementation, you would need to track the server process and send an appropriate signal or API request to trigger a graceful shutdown.

	gl.Info("Stopping GNyx Gateway...")

	stopGwCommand := "pkill -f 'gnyx gateway up'"
	if initArgs != nil && initArgs.Files.ConfigFile != "" {
		stopGwCommand = "pkill -f 'gnyx gateway up -f " + initArgs.Files.ConfigFile + "'"
	}

	if err := executeCommand(stopGwCommand); err != nil {
		return gl.Errorf("failed to stop gateway: %v", err)
	}

	return nil
}

// executeStatus checks the gateway status
func executeStatus(cmd *cobra.Command, args []string) error {
	var initArgs = &kbxMod.InitArgs{}

	// Build target URL

	initArgs.Runtime.Bind = kbxGet.EnvOr("KUBEX_GNYX_BIND", "0.0.0.0")
	initArgs.Runtime.Port = kbxGet.EnvOr("KUBEX_GNYX_PORT", "5000")
	initArgs.Runtime.Host = net.JoinHostPort(initArgs.Runtime.Bind, initArgs.Runtime.Port)

	var targetAddress string
	if initArgs.Runtime.Bind == "0.0.0.0" {
		targetAddress = net.JoinHostPort("localhost", initArgs.Runtime.Port)
	} else {
		targetAddress = net.JoinHostPort(initArgs.Runtime.Bind, initArgs.Runtime.Port)
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
	gl.Infof("  gnyx gateway advise --mode exec --provider openai --model gpt-4o-mini --scorecard ./scorecard.json")
	gl.Infof("")
	gl.Infof("Available modes: exec, code, ops, community")
	gl.Infof("Available providers: openai, anthropic, gemini, groq")

	// TODO: Implement full advise functionality using cmdAdvise
	return gl.Errorf("advise command not fully implemented yet")
}

// executeCommand executes a shell command and returns the output or error
func executeCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return gl.Errorf("command execution failed: %v", err)
	}
	gl.Infof("Command output: %s", string(output))
	return nil
}

// loadEnv loads environment variables from a specified .env file
func loadEnv(configPath string) {
	// Initialize environment variables, set them inside environment
	if err := godotenv.Load(configPath); err != nil {
		gl.Warnf("No .env file found at %s, proceeding with existing environment variables", configPath)
	}
	gl.Infof("Environment variables loaded from %s", configPath)
}
