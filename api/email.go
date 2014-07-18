package api

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"text/template"
)

var (
	signupTmpl,
	recoverPassTmpl *template.Template
)

type Message struct {
	Subject string
	Body    []byte
	To      []string
	From    string
}

func (msg *Message) String() string {
	return fmt.Sprintf("To:%s\r\nFrom:%s\r\nSubject: %s\r\nBody:%s",
		strings.Join(msg.To, ", "),
		msg.From,
		msg.Subject,
		string(msg.Body),
	)
}

// Creates a new message from a template; message body set to rendered template
func MessageFromTemplate(subject string,
	to []string,
	from string,
	t *template.Template,
	data interface{}) (*Message, error) {

	msg := &Message{
		Subject: subject,
		To:      to,
		From:    from,
	}
	b := &bytes.Buffer{}
	if err := t.Execute(b, data); err != nil {
		return nil, err
	}
	msg.Body = b.Bytes()
	return msg, nil
}

type Mailer interface {
	Mail(*Message) error
}

type smtpMailer struct {
	smtp.Auth
	config *AppConfig
}

func (m *smtpMailer) Mail(msg *Message) error {
	return smtp.SendMail(m.config.SmtpHost+":25", m.Auth, msg.From, msg.To, msg.Body)
}

type fakeMailer struct{}

func (m *fakeMailer) Mail(msg *Message) error {
	log.Println(msg)
	return nil
}

func newSmtpMailer(config *AppConfig) Mailer {
	m := &smtpMailer{config: config}
	m.Auth = smtp.PlainAuth("", config.SmtpName, config.SmtpPassword, config.SmtpHost)
	return m
}

func NewMailer(config *AppConfig) Mailer {
	var mailer Mailer
	if config.SmtpName == "" {
		log.Println("WARNING: using fake mailer, messages will not be sent by SMTP. " +
			"Set SMTP_NAME and SMTP_PASSWORD in environment to enable.")
		mailer = &fakeMailer{}
	} else {
		mailer = newSmtpMailer(config)
	}
	signupTmpl = parseTemplate(config, "signup.tmpl")
	recoverPassTmpl = parseTemplate(config, "recover_pass.tmpl")
	return mailer
}