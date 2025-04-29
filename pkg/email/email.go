package email

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/EM-Stawberry/Stawberry/config"
	"gopkg.in/gomail.v2"
)

//go:generate mockgen -source=$GOFILE -destination=mock_email/mock_email.go -package=mock_email

type MailerService interface {
	Registered(userName string, userMail string)
	StatusUpdate(offerID uint, status string, userMail string)
	OfferReceived(offerID uint, userMail string)
	Stop(ctx context.Context)
}

type SmtpMailer struct {
	enabled bool
	ctx     context.Context
	ctxCanc context.CancelFunc
	dialer  *gomail.Dialer
	queue   chan *gomail.Message
	wg      sync.WaitGroup
	mutex   sync.Mutex
	stopped bool
}

func NewMailer(emailCfg *config.EmailConfig) MailerService {
	ctx, cancel := context.WithCancel(context.Background())

	m := &SmtpMailer{
		ctx:     ctx,
		ctxCanc: cancel,
	}

	if !emailCfg.Enabled {
		slog.Info("email notifications are disabled")
		return m
	}

	m.enabled = true
	m.ctx = ctx
	m.dialer = gomail.NewDialer(emailCfg.SmtpHost, emailCfg.SmtpPort, emailCfg.From, emailCfg.Password)
	m.queue = make(chan *gomail.Message, emailCfg.QueueSize)

	m.wg.Add(emailCfg.WorkerPool)
	for i := range emailCfg.WorkerPool {
		go m.worker(i + 1)
	}

	slog.Info("email notifications are enabled")

	return m
}

func (m *SmtpMailer) Stop(ctx context.Context) {
	if !m.enabled {
		return
	}
	slog.Info("mailer is stopping")

	m.mutex.Lock()
	m.stopped = true
	m.ctxCanc()
	close(m.queue)
	m.mutex.Unlock()

	workersDone := make(chan struct{}, 1)
	go func() {
		m.wg.Wait()
		workersDone <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		slog.Info("email workers forcefully stopped (timeout)")
	case <-workersDone:
		slog.Info("email queue emptied, mailer workers stopped")
	}
}

func (m *SmtpMailer) worker(i int) {
	defer m.wg.Done()
	for {
		select {
		case <-m.ctx.Done():
			for msg := range m.queue {
				if err := m.dialer.DialAndSend(msg); err != nil {
					slog.Error("failed to send email", "error", err)
				}
			}
			return

		case msg, ok := <-m.queue:
			if !ok {
				return
			}
			if err := m.dialer.DialAndSend(msg); err != nil {
				slog.Error("failed to send email", "error", err, "worker id", i)
			}
		}
	}
}

func (m *SmtpMailer) enqueue(msg *gomail.Message) {
	if !m.enabled {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.stopped {
		slog.Info("failed to enqueue email, mailer is stopped", "email", msg.GetHeader("Subject"))
		return
	}
	m.queue <- msg
}

func (m *SmtpMailer) StatusUpdate(offerID uint, status string, userMail string) {
	if !m.enabled {
		return
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.dialer.Username)
	msg.SetHeader("To", userMail)
	msg.SetHeader("Subject", fmt.Sprintf("Stawberry: Offer Status Update (ID %d)", offerID))
	msg.SetBody("text/plain", fmt.Sprintf("The status of your offer (%d) has been changed to: %s", offerID, status))

	m.enqueue(msg)
}

func (m *SmtpMailer) OfferReceived(offerID uint, userMail string) {
	if !m.enabled {
		return
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.dialer.Username)
	msg.SetHeader("To", userMail)
	msg.SetHeader("Subject", fmt.Sprintf("Stawberry: New Offer Received (ID %d)", offerID))
	msg.SetBody("text/plain", fmt.Sprintf("A new offer (%d) has been received", offerID))

	m.enqueue(msg)
}

func (m *SmtpMailer) Registered(userName string, userMail string) {
	if !m.enabled {
		return
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.dialer.Username)
	msg.SetHeader("To", userMail)
	msg.SetHeader("Subject", "Welcome to Strawberry!")
	msg.SetBody("text/plain", fmt.Sprintf("Thank you for registering, %s.", userName))

	m.enqueue(msg)
}
