package jobhub

import (
	"testing"

	"github.com/cenkalti/backoff"
)

func TestRun(t *testing.T) {
	p := NewPipeline("Test Pipeline")

	jA, _ := p.AddJob(
		Job{
			Name: "A",
			Path: "./cmd/tests/simple_success",
		},
	)
	jB, _ := p.AddJob(
		Job{
			Name:    "B",
			Path:    "./cmd/tests/simple_failure",
			Retry:   5,
			Backoff: backoff.NewExponentialBackOff(),
		},
	)
	jC, _ := p.AddJob(
		Job{
			Name: "C",
			Path: "./cmd/tests/simple_success",
		},
	)
	jD, _ := p.AddJob(
		Job{
			Name: "D",
			Path: "./cmd/tests/simple_success",
		},
	)
	jE, _ := p.AddJob(
		Job{
			Name: "E",
			Path: "./cmd/tests/simple_success",
		},
	)
	jF, _ := p.AddJob(
		Job{
			Name: "F",
			Path: "./cmd/tests/simple_success",
		},
	)
	jG, _ := p.AddJob(
		Job{
			Name: "G",
			Path: "./cmd/tests/simple_success",
		},
	)
	jH, _ := p.AddJob(
		Job{
			Name: "H",
			Path: "./cmd/tests/simple_success",
		},
	)
	jI, _ := p.AddJob(
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

	if _, err := p.Run(); err == nil {
		t.Error("Run() should have returned an error")
	}
}

func TestTopologicalSort(t *testing.T) {
	p := NewPipeline("Test Pipeline")

	jA, _ := p.AddJob(Job{})
	jB, _ := p.AddJob(Job{})
	jC, _ := p.AddJob(Job{})
	jD, _ := p.AddJob(Job{})
	jE, _ := p.AddJob(Job{})
	p.AddJobDependency(jA, jB, jD)
	p.AddJobDependency(jB, jC, jE)
	p.AddJobDependency(jC, jD, jE)

	wantO1 := []int{5, 4, 3, 2, 1}
	wantO2 := []int{4, 5, 3, 2, 1}
	have, err := p.topologicalSort()
	if err != nil {
		t.Error(err)
	}

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
	p := NewPipeline("Test Pipeline")

	jA, _ := p.AddJob(Job{})
	jB, _ := p.AddJob(Job{})
	jC, _ := p.AddJob(Job{})
	jD, _ := p.AddJob(Job{})
	jE, _ := p.AddJob(Job{})
	p.AddJobDependency(jA, jB, jD)
	p.AddJobDependency(jB, jC, jE)
	p.AddJobDependency(jC, jD, jE)

	// test cycle detection
	p.AddJobDependency(jD, jB)

	if _, err := p.topologicalSort(); err == nil {
		t.Error("topologicalSort() should have returned an error")
	}
}
