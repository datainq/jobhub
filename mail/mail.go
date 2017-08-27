package mail

import (
	"bytes"
	"html/template"

	"github.com/datainq/jobhub"
	"gopkg.in/gomail.v2"
)

type Mail struct {
	From    string
	To      string
	Subject *template.Template
	Body    *template.Template

	Host     string
	Port     int
	Username string
	Password string
}

func (m Mail) executeTemplate(t *template.Template, data interface{}) (string, error) {
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (m Mail) SendStatus(pipelineStatus jobhub.PipelineStatus) error {
	message := gomail.NewMessage()
	message.SetHeader("From", m.From)
	message.SetHeader("To", m.To)
	s, err := m.executeTemplate(m.Subject, pipelineStatus)
	if err != nil {
		return err
	}
	b, err := m.executeTemplate(m.Body, pipelineStatus)
	if err != nil {
		return err
	}
	message.SetHeader("Subject", s)
	message.SetBody("text/html", b)
	d := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)
	if err := d.DialAndSend(message); err != nil {
		return err
	}
	return nil
}
