// Package types defines interfaces and types for AI providers
package types

import (
	"github.com/kubex-ecosystem/kbx"
)

type ServerCORS struct {
	AllowOrigins []string `yaml:"allow_origins"`
}

type ServerConfig struct {
	Addr  string     `yaml:"addr"`
	CORS  ServerCORS `yaml:"cors"`
	Debug bool       `yaml:"debug"`
}

type DefaultsConfig struct {
	TenantID                   string    `yaml:"tenant_id"`
	UserID                     string    `yaml:"user_id"`
	Byok                       string    `yaml:"byok"`
	NotificationProvider       *Provider `yaml:"notification_provider"`
	NotificationTimeoutSeconds int       `yaml:"notification_timeout_seconds"`
}

// Config holds the complete provider configuration
type Config struct {
	Server    *ServerConfig                `yaml:"server"`
	Defaults  *DefaultsConfig              `yaml:"defaults"`
	Providers map[string]LLMProviderConfig `yaml:"providers"`
}

// ------------------------ BOF: NEW types for providers ------------------------

// Provider interface defines the contract for AI providers
type Provider interface{ kbx.Provider }

type ProviderExt interface{ kbx.ProviderExt }

type ChatRequest = kbx.ChatRequest

type ChatChunk = kbx.ChatChunk

type Message = kbx.Message

// ProviderConfig holds configuration for a specific provider
type ProviderConfig = kbx.LLMProviderConfig

// LLMDevelopmentConfig holds development settings for LLM providers
type LLMDevelopmentConfig = kbx.LLMDevelopmentConfig

// LLMProviderConfig holds configuration for a specific LLM provider
type LLMProviderConfig struct {
	ProviderConfig `yaml:",inline" json:",inline" mapstructure:",squash"`
}

// LLMConfig holds the complete provider configuration - old Config
type LLMConfig = kbx.LLMConfig

// ------------------------- EOF: NEW types for providers ------------------------
