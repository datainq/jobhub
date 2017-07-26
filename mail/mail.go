package mail

import (
	"github.com/datainq/jobhub"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// type SMS struct {
// }

type EMail struct {
	From     string
	To       string
	Subject  string
	Host     string
	Port     int
	Username string
	Password string
	Log      logrus.FieldLogger
}

func (e *EMail) SendStatus(statusPipeline jobhub.PipelineStatus) {

	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", e.To)
	m.SetHeader("Subject", e.Subject)
	m.SetBody("text/html", "SOMETHING")
	d := gomail.NewDialer(e.Host, e.Port, e.Username, e.Password)
	if err := d.DialAndSend(m); err != nil {
		e.Log.Panic(err)
	}
}
