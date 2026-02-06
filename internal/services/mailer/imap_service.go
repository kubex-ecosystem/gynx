package mailer

import (
	"context"
	"os"
	"strings"

	"github.com/kubex-ecosystem/kbx"
	kbxTypes "github.com/kubex-ecosystem/kbx/types"

	kbxGet "github.com/kubex-ecosystem/kbx/get"
	kbxIs "github.com/kubex-ecosystem/kbx/is"
	imapclient "github.com/kubex-ecosystem/kbx/mailing/imap"
	gl "github.com/kubex-ecosystem/logz"
)

// IMAPConfig define os parâmetros para conexão IMAP.
type IMAPConfig = kbx.MailConnection

// IMAPService encapsula o acesso IMAP opcional.
type IMAPService struct {
	*IMAPConfig
	cfg *imapclient.Config
}

// FetchUnread retorna mensagens não lidas usando a configuração carregada.
func (s *IMAPService) FetchUnread(ctx context.Context) ([]*imapclient.Message, error) {
	if s == nil {
		return nil, gl.Errorf("IMAP service not initialized")
	}
	if s.cfg == nil {
		if kbxIs.Safe(s.IMAPConfig, false) {
			return nil, gl.Errorf("IMAP not configured")
		} else {
			s.cfg = &imapclient.Config{
				MailGeneralConfig:  kbxGet.ValOrType(s.MailGeneralConfig, kbxTypes.MailGeneralConfig{}),
				MailAuthConfig:     kbxGet.ValOrType(s.MailAuthConfig, kbxTypes.MailAuthConfig{}),
				MailProtocolConfig: kbxGet.ValOrType(s.MailProtocolConfig, kbxTypes.MailProtocolConfig{}),
			}
		}
	}
	return imapclient.FetchUnread(ctx, s.cfg)
}

// NewIMAPServiceFromPath tenta carregar a config IMAP a partir de um caminho.
// Se o caminho estiver vazio, não existir ou faltar host/user/pass, retorna (nil, nil).
func NewIMAPServiceFromPath(path string) (*IMAPService, error) {
	mailerConfig, err := kbx.LoadConfigOrDefault[kbx.MailConfig](path, true)
	if err != nil {
		return nil, err
	}
	if mailerConfig == nil {
		return nil, nil
	}

	for _, conn := range mailerConfig.Connections {
		if len(conn.Provider) == 0 ||
			len(conn.Host) == 0 ||
			len(conn.User) == 0 ||
			len(conn.Pass) == 0 ||
			!strings.EqualFold(conn.Provider, "imap") {
			continue
		}
		if strings.EqualFold(conn.Protocol, "imap") {
			return &IMAPService{
				IMAPConfig: &conn,
				cfg:        &conn,
			}, nil
		}
	}

	return nil, nil
}

// TryIMAPFromEnv tenta construir o serviço usando IMAP_CONFIG_FILE ou um fallback.
func TryIMAPFromEnv(fallbackPath string) (*IMAPService, error) {
	cfgPath := os.ExpandEnv(os.Getenv("IMAP_CONFIG_FILE"))
	if cfgPath == "" {
		cfgPath = fallbackPath
	}
	service, err := NewIMAPServiceFromPath(cfgPath)
	if err != nil {
		return nil, err
	}
	if service == nil {
		gl.Debug("IMAP not configured; skipping IMAP service init")
	}
	return service, nil
}
