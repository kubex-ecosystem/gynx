package contacts

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	mailer "github.com/kubex-ecosystem/gnyx/internal/services/mailer"
	ct "github.com/kubex-ecosystem/gnyx/internal/types"
	gl "github.com/kubex-ecosystem/logz"
)

func sendEmail(cc *ContactController, form ct.ContactForm) error {
	if cc == nil || cc.sender == nil || cc.smtpCfg == nil {
		return gl.Errorf("contact controller not properly configured")
	}

	from := cc.smtpCfg.User
	to := []string{cc.smtpCfg.User}
	subject := "PROFILE PAGE - New contact form submission"
	body := fmt.Sprintf("Name: %s\nEmail: %s\nMessage: %s", form.Name, form.Email, form.Message)

	msg := &mailer.EmailMessage{
		From:    from,
		To:      to,
		Subject: subject,
		Text:    body,
	}

	gl.Log("info", fmt.Sprintf("Sending email contact from %s to %s", form.Email, cc.smtpCfg.User))

	if err := cc.sender.Send(msg); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to send email via configured provider: %v", err))
		return err
	}

	gl.Log("success", "Email sent successfully")
	return nil
}

func sendEmailWithTimeout(cc *ContactController, form ct.ContactForm) error {
	if cc.sender == nil {
		return gl.Errorf("mailer not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Timeout definido
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- sendEmail(cc, form)
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() != nil {
			gl.Log("error", fmt.Sprintf("Timeout error: %v", ctx.Err().Error()))
			return errors.New("error: " + ctx.Err().Error())
		}
	case err := <-errChan:
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error sending email: %v", err.Error()))
			return err // Falha ao enviar
		}
	}

	gl.Log("success", "Email sent successfully within timeout")
	return nil // Sucesso no envio
}

func sendEmailWithRetry(cc *ContactController, form ct.ContactForm, attempts int) error {
	var err error
	for attemptsCounter := 0; attemptsCounter < attempts; attemptsCounter++ {
		err = sendEmailWithTimeout(cc, form)
		if err == nil {
			gl.Log("success", fmt.Sprintf("Email sent successfully after %d attempt(s)", attemptsCounter+1))
			return nil // Sucesso
		}
		// Implementa uma estratégia de retry exponencial:
		randomDelay := time.Duration(math.Pow(2, float64(attemptsCounter))) * time.Second
		time.Sleep(randomDelay)
	}
	return gl.Errorf("failed to send email after %d attempts: %v", attempts, err)
}
