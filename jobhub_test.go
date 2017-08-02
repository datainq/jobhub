package jobhub

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestAddJobDependency(t *testing.T) {
	retryCount := 1337
	p := NewPipeline()
	p.Name = "Example pipeline"
	logrus.SetLevel(logrus.DebugLevel)
	jA := p.AddJob(
		Job{
			Name: "A",
			Path: "./cmd/tests/simple_success",
		},
	)
	jB := p.AddJob(
		Job{
			Name:  "B",
			Path:  "./cmd/tests/simple_failure",
			Retry: retryCount,
		},
	)
	jC := p.AddJob(
		Job{
			Name: "C",
			Path: "./cmd/tests/simple_success",
		},
	)
	jD := p.AddJob(
		Job{
			Name: "D",
			Path: "./cmd/tests/simple_success",
		},
	)
	jE := p.AddJob(
		Job{
			Name: "E",
			Path: "./cmd/tests/simple_success",
		},
	)
	p.AddJobDependency(jA, jB, jD)
	p.AddJobDependency(jB, jC, jE)
	p.AddJobDependency(jC, jD, jE)

	status := p.Run()

	// test for retry number
	// B is the third job in queue
	if len(status.JobStatus[3].Statuses) != retryCount {
		t.Errorf("Wrong retry count")
	}
}
