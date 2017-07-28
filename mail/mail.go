package mail

import (
	"bytes"
	"html/template"
	"io/ioutil"

	"github.com/datainq/jobhub"
	"github.com/sirupsen/logrus"
)

// type SMS struct {
// }

type EMail struct {
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

func (e *EMail) SendStatus(pipelineStatus jobhub.PipelineStatus) {
	body, err := parseTemplate(e.TemplatePath, pipelineStatus)
	if err != nil {
		e.Log.Panic(err) //TODO(amwolff) do something else than panic
	}

	f := []byte(body)
	err = ioutil.WriteFile("./mail.html", f, 0644)
	if err != nil {
		e.Log.Panic(err)
	}
	// m := gomail.NewMessage()
	// m.SetHeader("From", e.From)
	// m.SetHeader("To", e.To)
	// m.SetHeader("Subject", e.Subject)
	// m.SetBody("text/html", body)
	// d := gomail.NewDialer(e.Host, e.Port, e.Username, e.Password)
	// if err = d.DialAndSend(m); err != nil {
	// 	e.Log.Panic(err) //TODO(amwolff) do something else than panic
	// }
}
