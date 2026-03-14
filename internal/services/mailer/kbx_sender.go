package mailer

import (
	"context"
	"fmt"
	"strings"

	"github.com/kubex-ecosystem/kbx"
	"github.com/kubex-ecosystem/kbx/mailing"
	"github.com/kubex-ecosystem/kbx/mailing/templates"
	"github.com/kubex-ecosystem/kbx/types"
)

// KBXSender adapta a interface MailSender para o mailer do kbx.
type KBXSender struct {
	*mailing.Mailer `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
}

// NewKBXSenderFromPath carrega a configuração SMTP/Retry a partir de um arquivo suportado pelo mapper.
// Aceita JSON/YAML/TOML/XML conforme contratos do Mapper.
func NewKBXSenderFromPath(cfgPath string) (*KBXSender, error) {
	mailConfig, err := kbx.LoadConfigOrDefault[kbx.MailConfig](cfgPath, true)
	if err != nil {
		return nil, fmt.Errorf("failed to load MailConfig: %v", err)
	}
	if mailConfig == nil {
		return nil, fmt.Errorf("mail config is nil")
	}
	if len(mailConfig.Connections) == 0 {
		return nil, fmt.Errorf("mail config has no connections")
	}

	validSMTP := make([]types.MailConnection, 0, len(mailConfig.Connections))
	for _, conn := range mailConfig.Connections {
		if !isValidSMTPConnection(conn) {
			continue
		}
		validSMTP = append(validSMTP, conn)
	}
	if len(validSMTP) == 0 {
		return nil, fmt.Errorf("mail config has no valid smtp connections")
	}
	mailConfig.Connections = validSMTP
	m := mailing.NewMailer(mailConfig)
	return &KBXSender{Mailer: m}, nil
}

// Send envia a mensagem usando o mailer do kbx.
func (s *KBXSender) Send(msg *EmailMessage) error {
	if s == nil || s.Mailer == nil {
		return fmt.Errorf("mailer not initialized")
	}
	if msg == nil {
		return fmt.Errorf("email message is nil")
	}
	req := &mailing.MailRequest{
		Name:        msg.Name,
		From:        msg.From,
		To:          msg.To,
		Cc:          msg.Cc,
		Bcc:         msg.Bcc,
		Subject:     msg.Subject,
		HTML:        msg.HTML,
		Text:        msg.Text,
		Attachments: msg.Attachments,
	}
	if conn := s.Mailer.GetSMTPConnection(); conn != nil {
		if strings.TrimSpace(conn.From) != "" {
			req.From = strings.TrimSpace(conn.From)
		}
		if strings.TrimSpace(req.Name) == "" && strings.TrimSpace(conn.Name) != "" {
			req.Name = strings.TrimSpace(conn.Name)
		}
	}
	return s.Mailer.Send(context.Background(), req)
}

// SendTemplate auxilia convites e outros fluxos internos que já possuem loader de template embedado.
func (s *KBXSender) SendTemplate(ctx context.Context, loader templates.TemplateLoader, name, to, subject, from string, data any) error {
	if s == nil || s.Mailer == nil {
		return fmt.Errorf("mailer not initialized")
	}
	return s.Mailer.SendTemplate(ctx, loader, name, data, to, subject, from)
}

func isValidSMTPConnection(conn types.MailConnection) bool {
	if !strings.EqualFold(strings.TrimSpace(conn.Protocol), "smtp") && strings.TrimSpace(conn.Protocol) != "" {
		return false
	}
	if strings.TrimSpace(conn.Provider) == "" {
		return false
	}
	if strings.TrimSpace(conn.Host) == "" {
		return false
	}
	if strings.TrimSpace(conn.User) == "" {
		return false
	}
	if strings.TrimSpace(conn.Pass) == "" {
		return false
	}
	if conn.Port <= 0 {
		return false
	}
	return true
}
