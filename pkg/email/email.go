package email

import (
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/EM-Stawberry/Stawberry/config"
)

type MailerService interface {
	Registered(userName string, userMail string) error
	StatusUpdate(offerID uint, status string, userMail string) error
	OfferReceived(offerID uint, userMail string) error
}

type SmtpMailer struct {
	enabled  bool
	from     string
	smtpAddr string
	auth     smtp.Auth
}

func NewMailer(emailCfg *config.EmailConfig) MailerService {
	m := &SmtpMailer{}

	if !emailCfg.Enabled {
		// todo: zap log "email notifications are disabled"
		slog.Info("email notifications are disabled")
		return m
	}

	m.enabled = true
	m.from = emailCfg.From
	m.smtpAddr = emailCfg.SmtpHost + ":" + emailCfg.SmtpPort
	m.auth = smtp.PlainAuth("", emailCfg.From, emailCfg.Password, emailCfg.SmtpHost)

	// todo: zap log "email notifications are enabled"
	slog.Info("email notifications are enabled")

	return m
}

// В будущем письма нужно будет приукрасить. На данный момент это однострочные затычки.

func (m *SmtpMailer) StatusUpdate(offerID uint, status string, userMail string) error {
	if !m.enabled {
		return nil
	}

	text := []byte(fmt.Sprintf("The status of your offer (%d) has been changed to: %s", offerID, status))
	err := smtp.SendMail(m.smtpAddr, m.auth, m.from, []string{userMail}, text)
	if err != nil {
		return err
	}
	return nil
}

func (m *SmtpMailer) OfferReceived(offerID uint, userMail string) error {
	if !m.enabled {
		return nil
	}

	text := []byte(fmt.Sprintf("The offer (%d) has been received", offerID))
	err := smtp.SendMail(m.smtpAddr, m.auth, m.from, []string{userMail}, text)
	if err != nil {
		return err
	}
	return nil
}

func (m *SmtpMailer) Registered(userName string, userMail string) error {
	if !m.enabled {
		return nil
	}

	text := []byte(fmt.Sprintf("Thank you for registering, %s.", userName))
	err := smtp.SendMail(m.smtpAddr, m.auth, m.from, []string{userMail}, text)
	if err != nil {
		return err
	}
	return nil
}
