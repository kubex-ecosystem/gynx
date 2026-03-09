// Package cli provides common functionality for command-line interface applications.
package cli

import "github.com/kubex-ecosystem/gnyx/internal/module/info"

func GetDescriptions(descriptionArg []string, hideBanner bool) map[string]string {
	return info.GetDescriptions(descriptionArg, hideBanner)
}

func ConcatenateExamples(examples []string) string {
	result := ""
	for i, example := range examples {
		if result == "" {
			result += example
		} else {
			result += "  " + example
		}
		if i < len(examples)-1 {
			result += "\n"
		}
	}
	return result
}

var ModuleInfo info.Manifest
