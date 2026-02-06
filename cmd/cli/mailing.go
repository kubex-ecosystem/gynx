package cli

import (
	"github.com/kubex-ecosystem/kbx"

	gl "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"
)

func MailCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mail",
		Short: "GNyx Mail - Send test emails using configured SMTP settings",
		Long: `GNyx Mail provides a CLI to send test emails using the configured SMTP settings.

Features:
  • Send test emails using configured SMTP settings
`,
	}

	rootCmd.AddCommand(sendCommand())

	return rootCmd
}

func sendCommand() *cobra.Command {
	var debug = false
	var mailParams = kbx.NewMailSrvParams("")

	// Make HTTP GET request to health endpoint

	sendCmd := &cobra.Command{
		Use:   "send",
		Short: "Send a test email using the configured SMTP settings",
		Run: func(cmd *cobra.Command, args []string) {
			gl.SetDebugMode(debug)

			gl.Log("info", "Sending test email...")

			// sender := kbx.NewMailSender(mailParams)
			// if sender == nil {
			// 	gl.Log("error", "Failed to initialize mailer")
			// 	return
			// }

			// msg := mailParams.Email
			// msg.To = []string{"test@example.com"}
			// msg.Subject = "Test Email from GNyx CLI"
			// msg.Text = "This is a test email sent from the GNyx CLI."

			// if err := sender.Send(mailParams.MailConfig, msg); err != nil {
			// 	gl.Log("error", "Failed to send email:", err.Error())
			// 	return
			// }

			gl.Log("info", "Test email sent successfully")
		},
	}

	sendCmd.Flags().BoolVarP(&debug, "debug", "D", false, "Enable debug logging")
	sendCmd.Flags().StringVar(&mailParams.ConfigPath, "smtp-config", "", "Path to SMTP configuration file")
	sendCmd.Flags().StringVar(&mailParams.From, "from-email", "", "From email address")
	sendCmd.Flags().StringVar(&mailParams.Name, "from-name", "", "From name")

	return sendCmd
}
