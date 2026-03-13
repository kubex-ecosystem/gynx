// Package module provides internal types and functions for the GNyx application.
package module

import (
	cc "github.com/kubex-ecosystem/gnyx/cmd/cli"
	vs "github.com/kubex-ecosystem/gnyx/internal/module/version"
	gl "github.com/kubex-ecosystem/logz"
	"github.com/spf13/cobra"

	"os"
	"strings"
)

type GNyx struct {
	parentCmdName string
	PrintBanner   bool
}

func (m *GNyx) Alias() string {
	return ""
}
func (m *GNyx) ShortDescription() string {
	return "A powerful AI powered application which provides a full suite of tools with Backend, CLI, GUI interfaces, and much more."
}
func (m *GNyx) LongDescription() string {
	return `GNyx is a powerful AI powered application which provides a full suite of tools with Backend, CLI, GUI interfaces, and much more.
It is designed to be a versatile and comprehensive solution for developers, DevOps teams, and AI practitioners. It offers a wide range of features including:

- Repository analysis and insights
- AI-driven code review and recommendations
- Integration with popular AI providers like OpenAI, Anthropic, Gemini, and Groq
- A user-friendly CLI for seamless interaction
- A robust backend for handling complex operations
- A GUI for visualizing data and managing configurations

Whether you're looking to optimize your codebase, gain insights into your repositories, or leverage AI for enhanced development workflows, GNyx has you covered.`
}
func (m *GNyx) Usage() string {
	return "gnyx [command] [args]"
}
func (m *GNyx) Examples() []string {
	return []string{
		"gnyx gateway serve -p '8888' -b '127.0.0.1' -f './config.yaml'",
		"gnyx daemon start -c './daemon_config.yaml'",
		"gnyx mail send -to '<recipient_email>' -sub 'Subject' -body 'Email body content'",
	}
}
func (m *GNyx) Active() bool {
	return true
}
func (m *GNyx) Module() string {
	return "gnyx"
}
func (m *GNyx) Execute() error {
	return m.Command().Execute()
}
func (m *GNyx) Command() *cobra.Command {
	gl.Log("debug", "Starting GNyx CLI...")

	var rtCmd = &cobra.Command{
		Use:     m.Module(),
		Aliases: []string{m.Alias()},
		Example: m.concatenateExamples(),
		Version: vs.GetVersion(),
		Annotations: cc.GetDescriptions([]string{
			m.LongDescription(),
			m.ShortDescription(),
		}, m.PrintBanner),
	}

	// Add subcommands to the root command
	rtCmd.AddCommand(cc.GatewayCmds())
	// rtCmd.AddCommand(cc.ServiceCmd())
	rtCmd.AddCommand(cc.NewDaemonCommand())
	// rtCmd.AddCommand(cc.NewGuiCommand())
	rtCmd.AddCommand(cc.MailCommand())
	rtCmd.AddCommand(cc.MetadataCommand())

	// Add more commands as needed
	rtCmd.AddCommand(vs.CliCommand())

	// Set usage definitions for the command and its subcommands
	setUsageDefinition(rtCmd)
	for _, c := range rtCmd.Commands() {
		setUsageDefinition(c)
		if !strings.Contains(strings.Join(os.Args, " "), c.Use) {
			if c.Short == "" {
				c.Short = c.Annotations["description"]
			}
		}
	}

	return rtCmd
}
func (m *GNyx) SetParentCmdName(rtCmd string) {
	m.parentCmdName = rtCmd
}
func (m *GNyx) concatenateExamples() string {
	examples := ""
	rtCmd := m.parentCmdName
	if rtCmd != "" {
		rtCmd = rtCmd + " "
	}
	for _, example := range m.Examples() {
		examples += rtCmd + example + "\n  "
	}
	return examples
}
