package mail

import (
	"encoding/json"

	"github.com/datainq/jobhub"
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
}

func (e *EMail) SendStatus(statusPipeline jobhub.PipelineStatus) {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", e.To)
	m.SetHeader("Subject", e.Subject)

	in := &statusPipeline
	body, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}

	m.SetBody("text/html", string(body))

	d := gomail.NewDialer(e.Host, e.Port, e.Username, e.Password)

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
