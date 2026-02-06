// Package cli provides common functionality for command-line interface applications.
package cli

import "github.com/kubex-ecosystem/gnyx/internal/module/info"

func GetDescriptions(descriptionArg []string, hideBanner bool) map[string]string {
	return info.GetDescriptions(descriptionArg, hideBanner)
}

var moduleInfo info.Manifest
