package main

import (
	"html/template"
	"log"

	"github.com/cenkalti/backoff"
	"github.com/datainq/jobhub"
	"github.com/datainq/jobhub/mail"
)

func catch(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	p := jobhub.NewPipeline("Example pipeline")
	jA, err := p.AddJob(
		jobhub.Job{
			Name: "A",
			Path: "./tests/simple_success",
		},
	)
	catch(err)
	jB, err := p.AddJob(
		jobhub.Job{
			Name:    "B",
			Path:    "./tests/simple_failure",
			Retry:   10,
			Backoff: backoff.NewExponentialBackOff(),
		},
	)
	catch(err)
	jC, err := p.AddJob(
		jobhub.Job{
			Name: "C",
			Path: "./tests/simple_success",
		},
	)
	catch(err)
	jD, err := p.AddJob(
		jobhub.Job{
			Name: "D",
			Path: "./tests/simple_success",
		},
	)
	catch(err)
	jE, err := p.AddJob(
		jobhub.Job{
			Name: "E",
			Path: "./tests/simple_success",
		},
	)
	catch(err)
	msg := mail.Mail{
		From: "From", // Sender's mail address
		To:   "To",   // Recipient's mail address

		// These fields are initialized with methods which return casts to template.Template
		// objects after parsing template files from given paths
		Subject: template.Must(template.ParseFiles("./default-templates/subject.html")),
		Body:    template.Must(template.ParseFiles("./default-templates/body.html")),

		Host:     "Host", // SMTP server address
		Port:     465,
		Username: "Username", // Server authentication credentials,
		Password: "Password", // the example values are self-explanatory
	}
	p.AddJobDependency(jA, jB, jD)
	p.AddJobDependency(jB, jC, jE)
	p.AddJobDependency(jC, jD, jE)
	if status, err := p.Run(); err != nil {
		if err = msg.SendStatus(status); err != nil {
			log.Fatal(err)
		}
	}
}
