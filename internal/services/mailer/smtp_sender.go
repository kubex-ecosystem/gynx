package mailer

import (
	"fmt"
	"net/smtp"

	"github.com/kubex-ecosystem/kbx"
	kbxTypes "github.com/kubex-ecosystem/kbx/types"
)

type MailServiceImpl struct {
	sender *MailSender `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
}

type MailSender interface {
	Send(msg *EmailMessage) error
}

func NewMailService(sender *MailSender) *MailServiceImpl {
	return &MailServiceImpl{sender: sender}
}

type SMTPSender struct {
	Host string    `json:"host" yaml:"host" xml:"host" toml:"host" mapstructure:"host"`
	Port int       `json:"port" yaml:"port" xml:"port" toml:"port" mapstructure:"port"`
	User string    `json:"username" yaml:"username" xml:"username" toml:"username" mapstructure:"username"`
	Pass string    `json:"password" yaml:"password" xml:"password" toml:"password" mapstructure:"password"`
	Auth smtp.Auth `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
}

type SMTPConfig = kbxTypes.MailConnection

func NewSMTPSender(configPath string) (*SMTPSender, error) {
	mailerConfig, err := kbx.LoadConfigOrDefault[kbx.MailConfig](configPath, true)
	if err != nil {
		return nil, err
	}
	if mailerConfig == nil {
		return nil, fmt.Errorf("mailer configuration is nil")
	}

	var config *kbxTypes.MailConnection
	for _, conn := range mailerConfig.Connections {
		if conn.Protocol == "smtp" || conn.Protocol == "" {
			config = &conn
			break
		}
	}
	auth := smtp.PlainAuth("", config.User, config.Pass, config.Host)
	return &SMTPSender{
		Host: config.Host,
		Port: config.Port,
		User: config.User,
		Pass: config.Pass,
		Auth: auth,
	}, nil
}
