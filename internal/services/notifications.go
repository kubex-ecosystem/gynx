package services

import (
	"context"
	"time"

	kbxTReg "github.com/kubex-ecosystem/kbx/tools/providers"
	kbxTypes "github.com/kubex-ecosystem/kbx/types"
)

// NotificationService handles sending notifications via different providers
type NotificationService struct {
	provider       string
	providerConfig string
	timeout        time.Duration
}

// NewNotificationService creates a new NotificationService with the given provider and timeout_seconds
func NewNotificationService(config *kbxTypes.SrvConfig) *NotificationService {
	if config == nil {
		return &NotificationService{provider: "", providerConfig: "", timeout: 0}
	}
	tout := config.Performance.TimeoutMS / 1000 // Convert milliseconds to seconds
	if tout == 0 {
		tout = 30 // Default timeout of 30 seconds if not specified
	}
	prv, err := kbxTReg.Load(config.Files.ProvidersConfig)
	if err != nil {
		return &NotificationService{provider: "", providerConfig: config.Files.ProvidersConfig, timeout: time.Duration(tout) * time.Second}
	}

	return &NotificationService{
		provider:       prv.ResolveProvider("default").Name(),
		providerConfig: config.Files.ProvidersConfig,
		timeout:        time.Duration(tout) * time.Second,
	}
}

// SendNotification sends a notification message using the configured provider
func (n *NotificationService) SendNotification(ctx context.Context, event kbxTypes.NotificationEvent) error {
	ctx, cancel := context.WithTimeout(ctx, n.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if n.provider == "" {
			return nil // No provider configured, nothing to do
		}
		lc := kbxTypes.NewLLMConfig(
			"",
			"",
			"",
			nil,
		)

		// Lógica temporária para evitar problemas de 'non implemented'
		p := kbxTReg.NewRegistry(
			&lc,
		).ResolveProvider(n.provider)

		if p == nil {
			return nil // Provider not found, nothing to do
		}

		return p.Notify(ctx, event)
	}
}

func (n *NotificationService) Name() string {
	return "NotificationService"
}
