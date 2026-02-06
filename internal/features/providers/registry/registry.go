// Package registry provides provider registration and resolution functionality.
package registry

import (
	"context"
	"fmt"
	"os"

	providers "github.com/kubex-ecosystem/gnyx/internal/types"
	gl "github.com/kubex-ecosystem/logz"
	"gopkg.in/yaml.v3"
)

// Registry manages provider registration and resolution
type Registry struct {
	cfg       providers.Config
	providers map[string]providers.Provider
}

// Load creates a new registry from a YAML configuration file
func Load(path string) (*Registry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, gl.Errorf("failed to read config file %s: %v", path, err)
	}

	var cfg providers.Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, gl.Errorf("failed to parse config file: %v", err)
	}

	r := &Registry{
		cfg:       cfg,
		providers: make(map[string]providers.Provider),
	}

	// Initialize providers based on configuration
	for name, pc := range cfg.Providers {
		switch pc.Type {
		case "openai":
			key := os.Getenv(pc.KeyEnv)
			if key == "" {
				gl.Log("warning", fmt.Sprintf("Skipping OpenAI provider '%s' - no API key found in %s", name, pc.KeyEnv))
				continue
			}
			p, err := NewOpenAIProvider(name, pc.BaseURL, key, pc.DefaultModel)
			if err != nil {
				return nil, gl.Errorf("failed to create OpenAI provider %s: %v", name, err)
			}
			r.providers[name] = p
		case "gemini":
			key := os.Getenv(pc.KeyEnv)
			if key == "" {
				gl.Log("warning", fmt.Sprintf("Skipping Gemini provider '%s' - no API key found in %s", name, pc.KeyEnv))
				continue
			}
			p, err := NewGeminiProvider(name, pc.BaseURL, key, pc.DefaultModel)
			if err != nil {
				return nil, gl.Errorf("failed to create Gemini provider %s: %v", name, err)
			}
			r.providers[name] = p
		case "anthropic":
			key := os.Getenv(pc.KeyEnv)
			if key == "" {
				gl.Log("warning", fmt.Sprintf("Skipping Anthropic provider '%s' - no API key found in %s", name, pc.KeyEnv))
				continue
			}
			p, err := NewAnthropicProvider(name, pc.BaseURL, key, pc.DefaultModel)
			if err != nil {
				return nil, gl.Errorf("failed to create Anthropic provider %s: %v", name, err)
			}
			r.providers[name] = p
		case "groq":

			key := os.Getenv(pc.KeyEnv)
			if key == "" {
				gl.Log("warning", fmt.Sprintf("Skipping Groq provider '%s' - no API key found in %s", name, pc.KeyEnv))
				continue
			}

			p, err := NewGroqProvider(name, pc.BaseURL, key, pc.DefaultModel)
			if err != nil {
				return nil, gl.Errorf("failed to create Groq provider %s: %v", name, err)
			}
			r.providers[name] = p
		case "openrouter":
			// TODO: Implement OpenRouter provider
			return nil, gl.Errorf("openrouter provider not yet implemented")
		case "ollama":
			// TODO: Implement Ollama provider
			return nil, gl.Errorf("ollama provider not yet implemented")
		default:
			return nil, gl.Errorf("unknown provider type: %s", pc.Type)
		}
	}

	return r, nil
}

// Resolve returns a provider by name
func (r *Registry) Resolve(name string) providers.Provider {
	return r.providers[name]
}

// ListProviders returns all available provider names
func (r *Registry) ListProviders() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// GetConfig returns the provider configuration
func (r *Registry) GetConfig() providers.Config {
	return r.cfg
}

func (r *Registry) ResolveProvider(name string) providers.Provider { return r.providers[name] }

func (r *Registry) Config() providers.Config { return r.cfg } // <- usado por /v1/providers

func (r *Registry) Chat(ctx context.Context, req providers.ChatRequest) (<-chan providers.ChatChunk, error) {
	p := r.ResolveProvider(req.Provider)
	if p == nil {
		return nil, gl.Errorf("provider '%s' not found", req.Provider)
	}
	return p.Chat(ctx, req)
}
func (r *Registry) Notify(ctx context.Context, event providers.NotificationEvent) error {
	p := r.ResolveProvider(event.Type)
	if p == nil {
		return gl.Errorf("provider '%s' not found", event.Type)
	}
	return p.Notify(ctx, event)
}

// /v1/chat/completions — SSE endpoints
