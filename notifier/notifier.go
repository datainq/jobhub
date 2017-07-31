package notifier

import (
	"bytes"
	"html/template"

	"github.com/datainq/jobhub"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// type SMS struct {
// }

type Mail struct {
	From         string
	To           string
	Subject      string
	Host         string
	Port         int
	Username     string
	Password     string
	TemplatePath string
	Log          logrus.FieldLogger
}

func parseTemplate(templatePath string, data interface{}) (string, error) {
	t := template.Must(template.ParseFiles(templatePath))
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (m *Mail) SendStatus(pipelineStatus jobhub.PipelineStatus) {
	if len(m.Subject) <= 0 {
		content := []string{
			pipelineStatus.Status.Code.String(),
			pipelineStatus.PipelineName,
			pipelineStatus.JobStatus[0].LastStatus.ExecutionStart.Format("2006-01-02 15:04:05"),
			pipelineStatus.Status.Runtime.String(),
		}
		for i, c := range content {
			if i == len(content)-1 {
				m.Subject += c
			} else {
				m.Subject += c + " - "
			}
		}
	}
	body, err := parseTemplate(m.TemplatePath, pipelineStatus)
	if err != nil {
		m.Log.Panic(err)
	}
	message := gomail.NewMessage()
	message.SetHeader("From", m.From)
	message.SetHeader("To", m.To)
	message.SetHeader("Subject", m.Subject)
	message.SetBody("text/html", body)
	d := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)
	if err = d.DialAndSend(message); err != nil {
		m.Log.Error(err)
	}
}
