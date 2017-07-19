package main

import "github.com/amwolff/jobhub"

func main() {
	p := jobhub.Pipeline{
		Name: "Example pipeline",
	}
	j0 := jobhub.Job{
		Name: "Failure",
		Path: "./tests/simple_failure",
	}
	j0 = p.AddJob(j0)
	j1 := jobhub.Job{
		Name: "Success",
		Path: "./tests/simple_success",
	}
	j1 = p.AddJob(j1)
	p.AddJobDependency(j1, j0)
	p.Run()
}
