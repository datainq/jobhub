package mail

import (
	"bytes"
	"html/template"

	"github.com/datainq/jobhub"
	"github.com/sirupsen/logrus"
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

	Log logrus.FieldLogger
}

func (m Mail) executeTemplate(t *template.Template, data interface{}) string {
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		m.Log.Panic(err)
	}
	return buf.String()
}

func (m Mail) SendStatus(pipelineStatus jobhub.PipelineStatus) {
	message := gomail.NewMessage()
	message.SetHeader("From", m.From)
	message.SetHeader("To", m.To)
	message.SetHeader("Subject", m.executeTemplate(m.Subject, pipelineStatus))
	message.SetBody("text/html", m.executeTemplate(m.Body, pipelineStatus))
	d := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)
	if err := d.DialAndSend(message); err != nil {
		m.Log.Error(err)
	}
}
