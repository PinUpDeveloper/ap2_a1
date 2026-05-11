package provider

import (
	"fmt"
	"log"
	"net/smtp"
)

type SMTPEmailSender struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func NewSMTPEmailSender(host, port, username, password, from string) *SMTPEmailSender {
	return &SMTPEmailSender{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (s *SMTPEmailSender) Send(msg EmailMessage) error {
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	body := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		s.From, msg.To, msg.Subject, msg.Body,
	)

	if err := smtp.SendMail(addr, auth, s.From, []string{msg.To}, []byte(body)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	log.Printf("[SMTPProvider] ✓ email sent to=%s subject=%q", msg.To, msg.Subject)
	return nil
}
