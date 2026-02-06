package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kubex-ecosystem/gnyx/internal/api/contacts"
	"github.com/kubex-ecosystem/gnyx/internal/config"
	"github.com/kubex-ecosystem/gnyx/internal/services/mailer"
	"github.com/kubex-ecosystem/gnyx/internal/types"
	"github.com/kubex-ecosystem/kbx"
	kbxTypes "github.com/kubex-ecosystem/kbx/types"
)

// RegisterContactRoutes cria o controller e registra os endpoints de contato, caso a config esteja disponível.
func RegisterContactRoutes(r *gin.RouterGroup, container types.IContainer) (gin.IRoutes, error) {
	if container == nil || container.Config() == nil {
		return r, fmt.Errorf("container or config is nil")
	}

	cfg, ok := container.Config().(*config.Config)
	if !ok || cfg == nil {
		return r, fmt.Errorf("invalid config type")
	}

	smtpCfg, err := loadSMTP(cfg.MailerConfigFilePath)
	if err != nil {
		return r, err
	}

	sender, err := mailer.NewKBXSenderFromPath(cfg.MailerConfigFilePath)
	if err != nil {
		return r, err
	}

	ctrl := contacts.NewContactController(sender, smtpCfg, nil)

	r.POST("/contact/handle", ctrl.HandleContact)
	r.GET("/contact", ctrl.GetContact)
	r.POST("/contact", ctrl.PostContact)
	r.GET("/contact/form", ctrl.GetContactForm)
	r.GET("/contact/form/:id", ctrl.GetContactFormByID)

	return r, nil
}

func loadSMTP(path string) (*kbxTypes.MailConnection, error) {
	mailerCfg, err := kbx.LoadConfigOrDefault[kbxTypes.MailConfig](path, true)
	if err != nil {
		return nil, err
	}
	if mailerCfg == nil {
		return nil, fmt.Errorf("mailer config is nil")
	}
	for _, conn := range mailerCfg.Connections {
		if conn.Protocol == "smtp" || conn.Protocol == "" {
			return &conn, nil
		}
	}
	return nil, fmt.Errorf("no SMTP connection found in mailer config")
}
