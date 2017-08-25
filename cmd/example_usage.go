package main

import (
	"html/template"

	"github.com/cenkalti/backoff"
	"github.com/datainq/jobhub"
	"github.com/datainq/jobhub/mail"
	"github.com/sirupsen/logrus"
)

func main() {
	p := jobhub.NewPipeline("Example pipeline", logrus.StandardLogger())
	logrus.SetLevel(logrus.DebugLevel)
	jA := p.AddJob(
		jobhub.Job{
			Name: "A",
			Path: "./tests/simple_success",
		},
	)
	jB := p.AddJob(
		jobhub.Job{
			Name:    "B",
			Path:    "./tests/simple_failure",
			Retry:   10,
			Backoff: backoff.NewExponentialBackOff(),
		},
	)
	jC := p.AddJob(
		jobhub.Job{
			Name: "C",
			Path: "./tests/simple_success",
		},
	)
	jD := p.AddJob(
		jobhub.Job{
			Name: "D",
			Path: "./tests/simple_success",
		},
	)
	jE := p.AddJob(
		jobhub.Job{
			Name: "E",
			Path: "./tests/simple_success",
		},
	)
	jF := p.AddJob(
		jobhub.Job{
			Name: "F",
			Path: "./tests/simple_success",
		},
	)
	jG := p.AddJob(
		jobhub.Job{
			Name: "G",
			Path: "./tests/simple_success",
		},
	)
	jH := p.AddJob(
		jobhub.Job{
			Name: "H",
			Path: "./tests/simple_success",
		},
	)
	jI := p.AddJob(
		jobhub.Job{
			Name: "I",
			Path: "./tests/simple_success",
		},
	)
	msg := mail.Mail{
		From: "From", //Sender's mail address
		To:   "To",   //Recipient's mail address

		// These fields are initialized with methods which return casts to template.Template
		// objects after parsing template files from given paths
		Subject: template.Must(template.ParseFiles("./default-templates/subject.html")),
		Body:    template.Must(template.ParseFiles("./default-templates/body.html")),

		Host:     "Host", //SMTP server address
		Port:     465,
		Username: "Username", //Server authentication credentials,
		Password: "Password", //the example values are self-explanatory
	}
	p.AddJobDependency(jA, jB, jD)
	p.AddJobDependency(jB, jC, jE, jF)
	p.AddJobDependency(jC, jD, jE)
	p.AddJobDependency(jF, jG)
	p.AddJobDependency(jG, jH)
	p.AddJobDependency(jH, jI)
	msg.SendStatus(p.Run())
}
