package main

import (
  "github.com/sirupsen/logrus"
  "github.com/amwolff/jobhub"
)

func main() {
  p := jobhub.Pipeline{
    Name: "My pipeline",
  }
  j0 := jobhub.Job{
    Name: "first",
    Path: "/bin/echo",
  }
  p.AddJob(j0)
  j1 := jobhub.Job{
    Name: "Second"
    Path: "/bin/echo",
  }
  
  p.AddJob(j1)
  p.AddJobDependency(j1, j0)
  p.Run()

  p := jobhub.Pipeline
  p.New


}