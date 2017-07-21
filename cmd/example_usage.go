package main

import (
	"github.com/amwolff/jobhub"
	"github.com/sirupsen/logrus"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	p := jobhub.NewPipeline()

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
	p.AddJobDependency(jA, jB, jD)
	p.AddJobDependency(jB, jC, jE)
	p.AddJobDependency(jC, jD, jE)
	//circular dependency
	//p.AddJobDependency(jD, jB)
	p.PrintDeps()
}
