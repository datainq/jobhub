package jobhub

import (
	"fmt"
	"testing"

	"github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
)

func TestRun(t *testing.T) {
	p := NewPipeline("My pipeline", logrus.StandardLogger())

	logrus.SetLevel(logrus.DebugLevel)

	jA := p.AddJob(
		Job{
			Name: "A",
			Path: "./cmd/tests/simple_success",
		},
	)
	jB := p.AddJob(
		Job{
			Name:    "B",
			Path:    "./cmd/tests/simple_failure",
			Retry:   -1,
			Backoff: backoff.NewExponentialBackOff(),
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
	jF := p.AddJob(
		Job{
			Name: "F",
			Path: "./cmd/tests/simple_success",
		},
	)
	jG := p.AddJob(
		Job{
			Name: "G",
			Path: "./cmd/tests/simple_success",
		},
	)
	jH := p.AddJob(
		Job{
			Name: "H",
			Path: "./cmd/tests/simple_success",
		},
	)
	jI := p.AddJob(
		Job{
			Name: "I",
			Path: "./cmd/tests/simple_success",
		},
	)
	p.AddJobDependency(jA, jB, jD)
	p.AddJobDependency(jB, jC, jE, jF)
	p.AddJobDependency(jC, jD, jE)
	p.AddJobDependency(jF, jG)
	p.AddJobDependency(jG, jH)
	p.AddJobDependency(jH, jI)

	p.Run()
}
func TestTopologicalSort(t *testing.T) {
	p := NewPipeline("My pipeline", logrus.StandardLogger())

	logrus.SetLevel(logrus.DebugLevel)

	jA := p.AddJob(Job{})
	jB := p.AddJob(Job{})
	jC := p.AddJob(Job{})
	jD := p.AddJob(Job{})
	jE := p.AddJob(Job{})
	p.AddJobDependency(jA, jB, jD)
	p.AddJobDependency(jB, jC, jE)
	p.AddJobDependency(jC, jD, jE)

	wantO1 := []int{5, 4, 3, 2, 1}
	wantO2 := []int{4, 5, 3, 2, 1}
	have := p.topologicalSort()

	for i, h := range have {
		if h != wantO1[i] {
			if h != wantO2[i] {
				t.Errorf("Sort does not match desired order (have %d vs want %d/%d)", have, wantO1, wantO2)
				return
			}
		}
	}
}

func TestTopologicalSortCycleDetection(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in f (%v)\n", r)
		}
	}()

	p := NewPipeline("My pipeline", logrus.StandardLogger())

	logrus.SetLevel(logrus.DebugLevel)

	jA := p.AddJob(Job{})
	jB := p.AddJob(Job{})
	jC := p.AddJob(Job{})
	jD := p.AddJob(Job{})
	jE := p.AddJob(Job{})
	p.AddJobDependency(jA, jB, jD)
	p.AddJobDependency(jB, jC, jE)
	p.AddJobDependency(jC, jD, jE)

	// test cycle detection
	p.AddJobDependency(jD, jB)

	p.topologicalSort()
	t.Error("The code did not panic")
}
