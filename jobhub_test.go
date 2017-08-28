package jobhub

import (
	"testing"

	"github.com/cenkalti/backoff"
)

func TestAddJob(t *testing.T) {
	p := Pipeline{}
	if _, err := p.AddJob(Job{}); err == nil {
		t.Errorf("AddJob() on %s should have returned an error", p.name)
	}

	q := NewPipeline("Test Pipeline")
	j, err := q.AddJob(Job{
		Name:    "Test Job",
		Path:    "path/to/binary",
		Args:    []string{"-abc"},
		Retry:   10,
		Backoff: backoff.NewExponentialBackOff(),
	},
	)
	if err != nil {
		t.Errorf("AddJob() on %s returned an error", q.name)
	}
	if j.id != 1 || len(q.jobContainer) != 1 {
		t.Errorf("AddJob() on %s misfunctioned", q.name)
	}
	if _, err := q.AddJob(j); err == nil {
		t.Errorf("AddJob() on %s should have returned an error", q.name)
	}
}

func TestAddJobDependency(t *testing.T) {
	p := NewPipeline("Test Pipeline")
	jX, _ := p.AddJob(Job{})
	jY, _ := p.AddJob(Job{})
	jZ, _ := p.AddJob(Job{})
	if err := p.AddJobDependency(jX, jY, jZ); err != nil {
		t.Errorf("AddJobDependency() on %s returned an error", p.name)
	}

	q := NewPipeline("And Another Test Pipeline")
	if err := q.AddJobDependency(Job{}); err == nil {
		t.Errorf("AddJobDependency() on %s should have returned an error", q.name)
	}
	jB, _ := q.AddJob(Job{})
	if err := q.AddJobDependency(jB, Job{}); err == nil {
		t.Errorf("AddJobDependency() on %s should have returned an error", q.name)
	}
}

func TestTopologicalSort(t *testing.T) {
	p := NewPipeline("Test Pipeline")
	j := make([]Job, 12)
	for i := 0; i < 12; i++ {
		j[i], _ = p.AddJob(Job{})
	}
	// https://upload.wikimedia.org/wikipedia/commons/1/1f/Depth-first-tree.svg
	p.AddJobDependency(j[0], j[1], j[6], j[7])
	p.AddJobDependency(j[1], j[2], j[5])
	p.AddJobDependency(j[2], j[3], j[4])
	p.AddJobDependency(j[7], j[8], j[11])
	p.AddJobDependency(j[8], j[9], j[10])
	want := []int{4, 5, 3, 6, 2, 7, 10, 11, 9, 12, 8, 1}
	have, err := p.topologicalSort()
	if err != nil {
		t.Error(err)
	}
	for i, h := range have {
		if h != want[i] {
			t.Errorf("Sort does not match desired order (have %d vs want %d)", have, want)
			return
		}
	}

	q := NewPipeline("Another Test Pipeline")
	jA, _ := q.AddJob(Job{})
	jB, _ := q.AddJob(Job{})
	jC, _ := q.AddJob(Job{})
	jD, _ := q.AddJob(Job{})
	jE, _ := q.AddJob(Job{})
	q.AddJobDependency(jA, jB, jD)
	q.AddJobDependency(jB, jC, jE)
	q.AddJobDependency(jC, jD, jE)
	q.AddJobDependency(jC, jD, jE) // this is on purpose
	q.AddJobDependency(jC, jD, jE)
	want = []int{4, 5, 3, 2, 1}
	have, err = q.topologicalSort()
	if err != nil {
		t.Error(err)
	}
	for i, h := range have {
		if h != want[i] {
			t.Errorf("Sort does not match desired order (have %d vs want %d)", have, want)
			return
		}
	}

	q.AddJobDependency(jD, jB)
	if _, err := q.topologicalSort(); err == nil {
		t.Errorf("topologicalSort() on %s should have returned an error", q.name)
	}
}

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
			Retry:   1,
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
	p.AddJobDependency(jB, jC, jE)
	p.AddJobDependency(jC, jD, jE)
	p.AddJobDependency(jF, jG)
	p.AddJobDependency(jG, jH)
	p.AddJobDependency(jH, jI)
	if _, err := p.Run(); err == nil {
		t.Errorf("Run() on %s should have returned an error", p.name)
	}
}
