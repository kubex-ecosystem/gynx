// Package bootstrap initializes and manages service providers based on configuration and environment variables.
package bootstrap

import (
	"embed"
	"io/fs"
	"path/filepath"
)

//go:embed all:services
var ServiceFiles embed.FS

func GetServices(name string) ([]string, error) {
	if name == "" {
		name = "*"
	}

	/*
		Read the directory entries from the embedded filesystem.
		All names must follow this structure:
		.Type-.-.Kind-.-.Environment-.-.Platform.template

		Where:
			- .Command: Type of service (e.g., gateway, database) - Command represents any function of this module offered as a service or the module itself without a specific function fixed (as a module).
			- .Type: Specific service type (e.g., service, runtime, container) - Type represents the way this service is delivered and executed, like a system service unit, a script, or a Dockerfile.
			- .Output: Container of the service (e.g., system, document) - Output represents the target environment or container for the service, such as a system service, a wrapper script, or a LookAtni document.
			- .Platform: Target platform (e.g., linux, windows, darwin) - Platform indicates the operating system for which this service is intended. Always inferred to match the current OS.
			- .template: File extension indicating it's a template file - Template files are used to generate actual service files based on the current environment and configuration.

		Examples:
			- [gateway]-[service]-[system]-[linux] 					- Represents a system service for a specific function on Linux.
			- [database]-[service]-[document]-[linux] 			- Represents a Blob LookAtni document for a specific function on Linux.
			- [function]-[service]-[document]-[linux] 			- Represents a Blob LookAtni document with a full service context for a specific function on Linux, unit and wrapper combined.
			- [function]-[runtime]-[linux] 									- Represents a runtime script for a specific function on Linux.
			- [function]-[container]-[document]-[linux] 		- Represents a Dockerfile template for a specific function on Linux, encapsulated in a LookAtni document with full context.
			- [function]-[dockerfile]-[system]-[linux] 			- Represents a Dockerfile ready to be used as a system service through a wrapper on Linux.
			- [function]-[dockerfile]-[wrapper]-[linux] 		- Represents a Dockerfile wrapper script for a specific function on Linux.
			- [function]-[dockerfile]-[document]-[linux]		- Represents a Dockerfile document with a full service context for a specific function on Linux, unit and wrapper combined.
			- [function]-[runtime]-[system]-[linux] 				- Represents a executable script ready to run as a system available command

		So, the CLI sends any part or a pattern to match against these names and retrieve the respective templates,
		we will infer the current platform, always, to avoid mismatches and install services incorrectly.

		Available Inputs:
			- Name: Command
				Description: Represents any function of this module offered as a service.
				Example: gateway
				Required: yes

			- Name: Type
				Description: Specific service type (e.g., service, script, dockerfile) - Optional
				Example: service
				Default: "script"
				Required: no

			- Name: Context
				Description: Container of the service (e.g., system, wrapper, lookatni) - Optional
				Example: system
				Default: "lookatni"
				Required: no

	*/

	entries, err := fs.ReadDir(ServiceFiles, "services")
	if err != nil {
		return nil, err
	}

	var services []string
	for _, entry := range entries {
		if match, _ := filepath.Match(name, entry.Name()); match {
			services = append(services, entry.Name())
		}
	}

	return services, nil
}
