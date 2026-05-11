package provider

type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

type EmailSender interface {
	Send(msg EmailMessage) error
}
