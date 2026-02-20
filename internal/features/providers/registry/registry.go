// Package registry provides provider registration and resolution functionality.
package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kbxMod "github.com/kubex-ecosystem/gnyx/internal/module/kbx"
	providers "github.com/kubex-ecosystem/gnyx/internal/types"
	kbxGet "github.com/kubex-ecosystem/kbx/get"
	kbxIs "github.com/kubex-ecosystem/kbx/is"
	kbxTools "github.com/kubex-ecosystem/kbx/tools"
	gl "github.com/kubex-ecosystem/logz"
)

// Registry manages provider registration and resolution
type Registry struct {
	cfg       providers.Config
	providers map[string]providers.Provider
}

// Load creates a new registry from a YAML configuration file
func Load(path string) (*Registry, error) {
	var rg = &Registry{
		cfg:       providers.Config{},
		providers: make(map[string]providers.Provider),
	}
	path = strings.TrimSpace(filepath.Clean(strings.ToValidUTF8(path, "")))

	if len(path) == 0 {
		gl.Warn("No provider config path specified. AI services will be unavailable.")
		return rg, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		gl.Warnf("Provider config file not found at %s. AI services will be unavailable.", path)
		return rg, nil
	} else if err != nil {
		return nil, gl.Errorf("error checking provider config file at %s: %v", path, err)
	}

	gl.Debugf("Loading provider configuration from %s", path)

	cfgMapper := kbxTools.NewEmptyMapperType[providers.Config](path)
	cfg, err := cfgMapper.DeserializeFromFile(filepath.Ext(path)[1:])
	if err != nil {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, gl.Errorf("failed to read config file %s: %v", path, err)
		}
		if len(b) == 0 {
			gl.Warnf("Provider config file at %s is empty. AI services will be unavailable.", path)
			return rg, nil
		}
		cfg, err := cfgMapper.Deserialize(b, filepath.Ext(path)[1:])
		if err != nil {
			return nil, gl.Errorf("failed to deserialize provider config from %s: %v", path, err)
		}
		if cfg == nil {
			gl.Warnf("Provider config file at %s is empty after deserialization. AI services will be unavailable.", path)
			return nil, gl.Errorf("provider config deserialized to nil from %s", path)
		}
	}

	rg = &Registry{
		cfg:       *kbxGet.ValOrType(cfg, &providers.Config{}),
		providers: make(map[string]providers.Provider),
	}

	// Initialize providers based on configuration
	for name, pc := range cfg.Providers {
		tp := strings.TrimSpace(strings.ToLower(strings.ToValidUTF8(pc.Type, "")))

		switch tp {
		case "openai":
			key := kbxGet.EnvOr(pc.KeyEnv, kbxGet.ValueOrIf(kbxIs.Map[string, string](cfg.Providers[name]), kbxGet.EnvOr(cfg.Providers[name].KeyEnv, kbxGet.EnvOr(kbxMod.DefaultLLMOpenAIKeyEnv, "")), ""))
			if key == "" {
				gl.Log("warning", fmt.Sprintf("Skipping OpenAI provider '%s' - no API key found in %s", name, pc.KeyEnv))
				continue
			}
			p, err := NewOpenAIProvider(name, pc.BaseURL, key, pc.DefaultModel)
			if err != nil {
				return nil, gl.Errorf("failed to create OpenAI provider %s: %v", name, err)
			}
			rg.providers[name] = p
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
			rg.providers[name] = p
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
			rg.providers[name] = p
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
			rg.providers[name] = p
		case "openrouter":
			// TODO: Implement OpenRouter provider
			return nil, gl.Errorf("openrouter provider not yet implemented")
		case "ollama":
			// TODO: Implement Ollama provider
			return nil, gl.Errorf("ollama provider not yet implemented")
		default:
			return nil, gl.Errorf("unknown provider type: %s", tp)
		}
	}

	return rg, nil
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
