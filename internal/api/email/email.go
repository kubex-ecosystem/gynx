// Package email fornece controladores API para acesso a e-mails via IMAP.
package email

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kubex-ecosystem/gnyx/internal/services/mailer"
	"github.com/kubex-ecosystem/kbx"
	gl "github.com/kubex-ecosystem/logz"
)

// AttachmentDTO descreve um anexo para consumo do FE.
type AttachmentDTO = kbx.MailAttachment

// EmailDTO descreve uma mensagem IMAP simplificada.
type EmailDTO = kbx.Email

// Controller encapsula o acesso IMAP já configurado.
type Controller struct {
	svc *mailer.IMAPService
}

func NewController(svc *mailer.IMAPService) *Controller {
	return &Controller{svc: svc}
}

// Register vincula as rotas de e-mail se o serviço existir.
func Register(r *gin.RouterGroup, ctl *Controller) gin.IRoutes {
	if ctl == nil || ctl.svc == nil {
		gl.Log("info", "IMAP service not configured; skipping /email endpoints")
		return r
	}
	r.GET("/email/unread", ctl.ListUnread)
	return r
}

// ListUnread retorna mensagens não lidas (sem quebrar se IMAP estiver indisponível).
func (c *Controller) ListUnread(ctx *gin.Context) {
	if c == nil || c.svc == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "disabled", "message": "IMAP not configured"})
		return
	}

	msgs, err := c.svc.FetchUnread(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	out := make([]*EmailDTO, 0, len(msgs))
	for _, m := range msgs {
		attachments := make([]AttachmentDTO, 0, len(m.Attachments))
		for _, att := range m.Attachments {
			attachments = append(attachments, AttachmentDTO{
				Filename: att.Filename,
				Size:     len(att.Data),
				Data:     []byte(base64.StdEncoding.EncodeToString(att.Data)),
			})
		}
		out = append(out, &EmailDTO{
			UID:         m.UID,
			From:        m.From,
			Subject:     m.Subject,
			Text:        m.Text,
			Attachments: attachments,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   out,
	})
}
