// Package registry provides provider registration and resolution functionality.
package registry

import (
	kbxTReg "github.com/kubex-ecosystem/kbx/tools/providers"
)

// Registry manages provider registration and resolution
type Registry = kbxTReg.Registry

// Load creates a new registry from a YAML configuration file
func Load(path string) (*Registry, error) { return kbxTReg.Load(path) }
