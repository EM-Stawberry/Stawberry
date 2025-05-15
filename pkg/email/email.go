package email

import (
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/spf13/viper"
)

var enableMail bool

var (
	from     string
	password string
	smtpHost string
	smtpPort string
	smtpAddr string
	auth     smtp.Auth
)

func SetupEmail(ok bool) {
	if !ok {
		return
	}
	enableMail = true

	slog.Info("email notifications are enabled")

	from = viper.GetString("FROM_EMAIL")
	password = viper.GetString("FROM_PASSWORD")
	smtpHost = viper.GetString("SMTP_HOST")
	smtpPort = viper.GetString("SMTP_PORT")
	smtpAddr = smtpHost + ":" + smtpPort
	auth = smtp.PlainAuth("", from, password, smtpHost)
}

func StatusUpdate(offerID uint, status string, userMail string) error {
	if !enableMail {
		return nil
	}

	text := []byte(fmt.Sprintf("The status of your offer (%d) has been changed to: %s", offerID, status))
	err := smtp.SendMail(smtpAddr, auth, from, []string{userMail}, text)
	if err != nil {
		return err
	}
	return nil
}

func OfferReceived(offerID uint, userMail string) error {
	if !enableMail {
		return nil
	}

	text := []byte(fmt.Sprintf("The offer (%d) has been received", offerID))
	err := smtp.SendMail(smtpAddr, auth, from, []string{userMail}, text)
	if err != nil {
		return err
	}
	return nil
}

func Registered(userName string, userMail string) error {
	if !enableMail {
		return nil
	}

	text := []byte(fmt.Sprintf("Thank you for registering, %s.", userName))
	err := smtp.SendMail(smtpAddr, auth, from, []string{userMail}, text)
	if err != nil {
		return err
	}
	return nil
}
