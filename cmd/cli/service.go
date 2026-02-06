package cli

import (
	_ "embed"
	// "fmt"
	// "os"
	// "os/exec"
	// "path/filepath"
	// "runtime"
	// "strings"

	// "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	"github.com/spf13/cobra"
)

// -------------------------------------------------------------------
// UNIVERSAL SERVICE MANAGER FOR ANY MODULE
// -------------------------------------------------------------------

func NewServiceCommand(moduleName, defaultCmd string, binPath string, defaultConfig string) *cobra.Command {
	// var initArgs = &kbx.InitArgs{}

	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage OS service for this Kubex module",
	}

	// cmd.AddCommand(cmdInstall(initArgs))
	// cmd.AddCommand(cmdUninstall(moduleName))
	// cmd.AddCommand(cmdStart(moduleName))
	// cmd.AddCommand(cmdStop(moduleName))
	// cmd.AddCommand(cmdStatus(moduleName))

	return cmd
}

// func systemctlCmd() string {
// 	if os.Geteuid() == 0 {
// 		return "systemctl"
// 	}
// 	return "systemctl --user"
// }

// // -------------------------------------------------------------------
// // INSTALL
// // -------------------------------------------------------------------

// func cmdInstall(initArgs *kbx.InitArgs) *cobra.Command {
// 	if initArgs == nil {
// 		initArgs = &kbx.InitArgs{}
// 	}

// 	cmd := &cobra.Command{
// 		Use:   "install",
// 		Short: "Install module as OS service",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			switch runtime.GOOS {
// 			case "linux":
// 				return installLinux(initArgs)
// 			case "windows":
// 				return installWindows(initArgs)
// 			default:
// 				return gl.Errorf("OS não suportado")
// 			}
// 		},
// 	}

// 	cmd.Flags().BoolVarP(&initArgs.Debug, "debug", "d", false, "Habilitar modo debug para o serviço")
// 	cmd.Flags().BoolVarP(&initArgs.FailFast, "fail-fast", "f", false, "Habilitar modo fail-fast para o serviço")
// 	cmd.Flags().BoolVarP(&initArgs.BatchMode, "batch-mode", "b", false, "Habilitar modo batch para o serviço")
// 	cmd.Flags().BoolVarP(&initArgs.NoColor, "no-color", "", false, "Desabilitar cores no log")
// 	cmd.Flags().BoolVarP(&initArgs.RootMode, "root-mode", "", false, "Habilitar modo root para o serviço")

// 	cmd.Flags().StringVarP(&initArgs.Name, "name", "", "localhost", "Nome do serviço")
// 	cmd.Flags().StringVarP(&initArgs.Host, "host", "", "localhost", "Host para o serviço escutar")
// 	cmd.Flags().StringVarP(&initArgs.Command, "command", "", "localhost", "Comando para o serviço executar")
// 	cmd.Flags().StringVarP(&initArgs.Subcommand, "subcommand", "", "localhost", "Subcomando para o serviço executar")
// 	cmd.Flags().StringVarP(&initArgs.ConfigFile, "config", "c", os.ExpandEnv(kbx.DefaultGNyxCAPath), "Caminho para o arquivo de configuração")
// 	cmd.Flags().StringVarP(&initArgs.EnvFile, "env", "e", "", "Caminho para o arquivo .env")
// 	cmd.Flags().StringVarP(&initArgs.LogFile, "log", "l", "", "Caminho para o arquivo de log")

// 	cmd.Flags().IntVarP(&initArgs.MaxProcs, "max-procs", "", 0, "Número máximo de processos")
// 	cmd.Flags().IntVarP(&initArgs.TimeoutMS, "timeout-ms", "", 0, "Timeout em milissegundos para operações")

// 	cmd.Flags().StringToStringVarP(&initArgs.EnvVars, "option", "o", nil, "Variável de ambiente adicional no formato CHAVE=VALOR (aceita múltiplas)")

// 	return cmd
// }

// func installLinux(initArgs *kbx.InitArgs) error {
// 	name := initArgs.Name
// 	defaultCmd := initArgs.Command
// 	binPath := moduleInfo.GetBin()
// 	defaultConfig := initArgs.ConfigFile

// 	isRoot := os.Geteuid() == 0 || initArgs.RootMode
// 	home, _ := os.UserHomeDir()

// 	var unitPath, wrapperPath, serviceUser string

// 	if isRoot {
// 		unitPath = fmt.Sprintf("/etc/systemd/system/%s.service", name)
// 		wrapperPath = "/usr/local/bin/kubex-svc"
// 		serviceUser = "appuser"
// 	} else {
// 		unitPath = fmt.Sprintf("%s/.config/systemd/user/%s.service", home, name)
// 		wrapperPath = fmt.Sprintf("%s/.local/bin/kubex-svc", home)
// 		serviceUser = os.Getenv("USER")
// 		os.MkdirAll(filepath.Dir(unitPath), 0755)
// 		os.MkdirAll(filepath.Dir(wrapperPath), 0755)
// 	}
// 	wrapper := []byte{}
// 	// write wrapper
// 	os.WriteFile(wrapperPath, []byte(wrapper), 0755)

// 	// compile systemd template
// 	unit := systemdTemplate
// 	unit = strings.ReplaceAll(unit, "{{MODULE_NAME}}", name)
// 	unit = strings.ReplaceAll(unit, "{{MODULE_BIN}}", binPath)
// 	unit = strings.ReplaceAll(unit, "{{MODULE_CONFIG}}", defaultConfig)
// 	unit = strings.ReplaceAll(unit, "{{MODULE_DEFAULT_CMD}}", defaultCmd)
// 	unit = strings.ReplaceAll(unit, "{{WRAPPER_PATH}}", wrapperPath)
// 	unit = strings.ReplaceAll(unit, "{{SERVICE_USER}}", serviceUser)

// 	// write unit
// 	os.WriteFile(unitPath, []byte(unit), 0644)

// 	// reload + enable
// 	exec.Command("sh", "-c", systemctlCmd()+" daemon-reload").Run()
// 	exec.Command("sh", "-c", systemctlCmd()+" enable "+name).Run()

// 	gl.Println("✓ Serviço Kubex instalado:", name)
// 	return nil
// }

// // -------------------------------------------------------------------
// // WINDOWS
// // -------------------------------------------------------------------

// func installWindows(args *kbx.InitArgs) error {
// 	name := args.Name
// 	binPath := moduleInfo.GetBin()
// 	config := args.ConfigFile

// 	os.MkdirAll("C:\\kubex", 0755)

// 	path := fmt.Sprintf("C:\\kubex\\install-%s.ps1", name)
// 	content := strings.ReplaceAll(ps1Template, "$ModuleName", name)
// 	content = strings.ReplaceAll(content, "$BinaryPath", binPath)
// 	content = strings.ReplaceAll(content, "$ConfigPath", config)

// 	os.WriteFile(path, []byte(content), 0644)

// 	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", path)
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	return cmd.Run()
// }

// // -------------------------------------------------------------------
// // START/STOP/STATUS
// // -------------------------------------------------------------------

// func cmdStart(name string) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "start",
// 		Short: "Start service",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return exec.Command("sh", "-c", systemctlCmd()+" start "+name).Run()
// 		},
// 	}
// }

// func cmdStop(name string) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "stop",
// 		Short: "Stop service",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return exec.Command("sh", "-c", systemctlCmd()+" stop "+name).Run()
// 		},
// 	}
// }

// func cmdStatus(name string) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "status",
// 		Short: "Show service status",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			return exec.Command("sh", "-c", systemctlCmd()+" status "+name).Run()
// 		},
// 	}
// }

// func cmdUninstall(name string) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "uninstall",
// 		Short: "Remove service",
// 		RunE: func(cmd *cobra.Command, args []string) error {

// 			exec.Command("sh", "-c", systemctlCmd()+" stop "+name).Run()
// 			exec.Command("sh", "-c", systemctlCmd()+" disable "+name).Run()

// 			os.Remove("/etc/systemd/system/" + name + ".service")
// 			os.Remove(os.Getenv("HOME") + "/.config/systemd/user/" + name + ".service")

// 			gl.Println("✓ Serviço removido:", name)
// 			return nil
// 		},
// 	}
// }
