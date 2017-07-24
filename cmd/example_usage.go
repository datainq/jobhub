package main

import (
	"github.com/datainq/jobhub"
	"github.com/sirupsen/logrus"
)

func main() {
	p := jobhub.NewPipeline()
	logrus.SetLevel(logrus.DebugLevel)
	jA := p.AddJob(
		jobhub.Job{
			Name: "A",
			Path: "./tests/simple_success",
		},
	)
	jB := p.AddJob(
		jobhub.Job{
			Name: "B",
			Path: "./tests/simple_success",
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
	p.AddJobDependency(jA, jB, jD)
	p.AddJobDependency(jB, jC, jE, jF)
	p.AddJobDependency(jC, jD, jE)
	p.AddJobDependency(jF, jG)
	p.AddJobDependency(jG, jH)
	p.AddJobDependency(jH, jI)
	p.Run()
	p.PrintDeps()
}
