// Package cli provides the daemon command for background service operations
package cli

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubex-ecosystem/gnyx/internal/app/daemon"
	"github.com/spf13/cobra"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
	gl "github.com/kubex-ecosystem/logz"
)

var (
	gnyxURL             string
	gnyxAPIKey          string
	autoScheduleEnabled bool
	scheduleCron        string
	notifyChannels      []string
	healthCheckInterval time.Duration
)

// NewDaemonCommand creates the daemon command
func NewDaemonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Start gnyx as background daemon service",
		Long: `Start the gnyx as a background daemon service.

The daemon provides:
• Automatic repository analysis scheduling
• Integration with KubeX AI Squad system
• Discord/WhatsApp/Email notifications
• Health monitoring and reporting
• Meta-recursivity coordination with lookatni/grompt

Examples:
  gnyx daemon --gnyx-url=http://localhost:5000 --gnyx-api-key=abc123
  gnyx daemon --auto-schedule --schedule-cron="0 2 * * *"
  gnyx daemon --notify-channels=discord,email`,
		RunE: runDaemon,
	}

	// GNyx Integration flags
	cmd.Flags().StringVar(&gnyxURL, "gnyx-url",
		getEnvOrDefault("KUBEX_GNYX_URL", "http://localhost:"+kbxGet.EnvOr("KUBEX_GNYX_PORT", "5000")),
		"GNyx backend URL")
	cmd.Flags().StringVar(&gnyxAPIKey, "gnyx-api-key",
		os.Getenv("KUBEX_GNYX_API_KEY"),
		"GNyx API key for authentication")

	// Scheduling flags
	cmd.Flags().BoolVar(&autoScheduleEnabled, "auto-schedule", false,
		"Enable automatic repository analysis scheduling")
	cmd.Flags().StringVar(&scheduleCron, "schedule-cron", "0 2 * * *",
		"Cron expression for automatic scheduling (default: daily at 2 AM)")

	// Notification flags
	cmd.Flags().StringSliceVar(&notifyChannels, "notify-channels",
		[]string{"discord"},
		"Notification channels (discord,email,webhook)")

	// Health monitoring flags
	cmd.Flags().DurationVar(&healthCheckInterval, "health-interval",
		5*time.Minute,
		"Health check interval")

	return cmd
}

func runDaemon(cmd *cobra.Command, args []string) error {
	// Validate required flags
	if gnyxAPIKey == "" {
		return gl.Errorf("--gnyx-api-key is required (or set GNYX_API_KEY env var)")
	}

	// Create daemon configuration
	config := daemon.DaemonConfig{
		GNyxURL:              gnyxURL,
		GNyxAPIKey:           gnyxAPIKey,
		AutoScheduleEnabled:  autoScheduleEnabled,
		ScheduleCron:         scheduleCron,
		NotificationChannels: notifyChannels,
		HealthCheckInterval:  healthCheckInterval,
	}

	// Create and start daemon
	d := daemon.NewAnalyzerDaemon(config)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start daemon
	if err := d.Start(); err != nil {
		return gl.Errorf("failed to start daemon: %v", err)
	}

	// Print startup information
	printDaemonInfo(config)

	// Wait for shutdown signal
	<-sigChan
	gl.Log("info", "Received shutdown signal, stopping daemon...")

	// Graceful shutdown
	d.Stop()
	gl.Log("info", "GNyx daemon stopped gracefully")

	return nil
}

func printDaemonInfo(config daemon.DaemonConfig) {
	gl.Log("info", "")
	gl.Log("info", "========================== Daemon Startup ============================")
	gl.Log("info", "🤖   GNYX DAEMON - Repository Intelligence Platform")
	gl.Log("info", "============================================================")
	gl.Log("info", "")
	gl.Infof("🏗️  GNyx Integration: %s", config.GNyxURL)
	gl.Infof("📅 Auto Schedule: %v", config.AutoScheduleEnabled)
	if config.AutoScheduleEnabled {
		gl.Infof(" (%s) ", config.ScheduleCron)
	}
	gl.Infof("🔔 Notifications: %v", config.NotificationChannels)
	gl.Infof("🏥 Health Checks: every %v", config.HealthCheckInterval)
	gl.Log("info", "")
	gl.Log("info", "📊 CAPABILITIES:")
	gl.Log("info", "   • Repository Intelligence Analysis")
	gl.Log("info", "   • DORA Metrics Collection")
	gl.Log("info", "   • Code Health Index (CHI)")
	gl.Log("info", "   • AI Impact Analysis")
	gl.Log("info", "   • Automated Scheduling")
	gl.Log("info", "   • Multi-channel Notifications")
	gl.Log("info", "   • KubeX AI Squad Integration")
	gl.Log("info", "   • Meta-recursivity Coordination")
	gl.Log("info", "")
	gl.Log("info", "🎯 INTEGRATION POINTS:")
	gl.Log("info", "   • GNyx Backend APIs")
	gl.Log("info", "   • Discord Webhooks")
	gl.Log("info", "   • Email Notifications")
	gl.Log("info", "   • GitHub Events")
	gl.Log("info", "   • Jira Workflows (planned)")
	gl.Log("info", "   • WakaTime Analytics (planned)")
	gl.Log("info", "")
	gl.Log("info", "🔄 META-RECURSIVITY:")
	gl.Log("info", "   • Coordinates with lookatni (analysis)")
	gl.Log("info", "   • Orchestrates grompt (improvement)")
	gl.Log("info", "   • Manages continuous optimization")
	gl.Log("info", "Daemon running... Press Ctrl+C to stop")
	gl.Log("info", "")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
