// Package registry provides provider registration and resolution functionality.
package registry

import (
	"os"
	"strings"

	config "github.com/kubex-ecosystem/gnyx/internal/config"
	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	kbxTReg "github.com/kubex-ecosystem/kbx/tools/providers"
)

// Registry manages provider registration and resolution.
type Registry = kbxTReg.Registry

// ResolvePath returns the effective providers config path for the current runtime.
func ResolvePath(cfg *config.ServerConfig) string {
	candidates := []string{}
	if cfg != nil {
		candidates = append(candidates, cfg.ProvidersConfig)
		candidates = append(candidates, cfg.Files.ProvidersConfig)
	}
	candidates = append(candidates,
		os.Getenv("KUBEX_GNYX_PROVIDERS_CONFIG_PATH"),
		kbxMod.DefaultProvidersConfig,
	)

	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		return os.ExpandEnv(candidate)
	}

	return os.ExpandEnv(kbxMod.DefaultProvidersConfig)
}

// LoadResolved loads the provider registry using the effective runtime config path
// and keeps both config fields in sync so later bootstrap stages don't drift.
func LoadResolved(cfg *config.ServerConfig) (*Registry, error) {
	path := ResolvePath(cfg)
	if cfg != nil {
		cfg.ProvidersConfig = path
		cfg.Files.ProvidersConfig = path
	}
	return kbxTReg.Load(path)
}

// Load creates a new registry from a YAML configuration file.
func Load(path string) (*Registry, error) { return kbxTReg.Load(path) }
