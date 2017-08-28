package jobhub

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cenkalti/backoff"
)

var pipelineIDPool = 1

type pipeline struct {
	name          string
	id            int
	nextJobID     int
	jobContainer  []Job
	jobByID       map[int]Job
	jobDependency map[int][]int
}

type Job struct {
	Name    string
	Path    string
	Args    []string
	Retry   int
	Backoff backoff.BackOff

	id         int
	pipelineID int
}

type PipelineStatus struct {
	PipelineName string
	Status       ExecutionStatus
	JobStatus    []JobStatus
}

type JobStatus struct {
	Job        Job
	JobID      int
	LastStatus ExecutionStatus
	Statuses   []ExecutionStatus
}

type ExecutionStatus struct {
	ExecutionStart time.Time
	Code           StatusCode
	Runtime        time.Duration
}

type StatusCode int

const (
	Created   StatusCode = 0
	Scheduled StatusCode = 1
	Failed    StatusCode = 2
	Succeeded StatusCode = 3
)

var statusDescription = map[StatusCode]string{
	Created:   "Created",
	Scheduled: "Scheduled",
	Failed:    "Failed",
	Succeeded: "Succeeded",
}

func (j JobStatus) LastExecutionStatus() *ExecutionStatus {
	if len(j.Statuses) > 0 {
		return &j.Statuses[len(j.Statuses)-1]
	}
	return nil
}

func (s StatusCode) String() string {
	return statusDescription[s]
}

func NewPipeline(name string) *pipeline {
	return &pipeline{
		name:          name,
		jobByID:       make(map[int]Job),
		jobDependency: make(map[int][]int),
	}
}

func nextIDPipeline() int {
	tempID := pipelineIDPool
	pipelineIDPool++
	return tempID
}

func (p *pipeline) nextIDJob() int {
	p.nextJobID++
	return p.nextJobID
}

func (p *pipeline) AddJob(job Job) (Job, error) {
	for _, j := range p.jobContainer {
		if j.id == job.id {
			return j, fmt.Errorf("%s (%d) in %s (%d) has already been added", j.Name, j.id, p.name, p.id)
		}
	}
	if p.id == 0 {
		p.id = nextIDPipeline()
	}
	job.pipelineID = p.id
	job.id = p.nextIDJob()
	p.jobContainer = append(p.jobContainer, job)
	p.jobByID[job.id] = job
	return job, nil
}

func (p *pipeline) AddJobDependency(job Job, deps ...Job) error {
	tempContainer := append(deps, job)
	if p.id == 0 {
		return fmt.Errorf("%s (%d) not initialized (i.e. has no jobs added)", p.name, p.id)
	}
	for _, j := range tempContainer {
		if j.pipelineID != p.id {
			return fmt.Errorf("%s (%d) does not belong to %s (%d)", j.Name, j.id, p.name, p.id)
		}
	}
	for _, d := range deps {
		p.jobDependency[job.id] = append(p.jobDependency[job.id], d.id)
	}
	return nil
}

func (p pipeline) topologicalSort() ([]int, error) {
	var (
		temporaryMark = map[int]bool{}
		permanentMark = map[int]bool{}
		acyclic       = true
		queue         []int
		visit         func(int)
	)

	visit = func(u int) {
		if permanentMark[u] {
			return
		} else if temporaryMark[u] {
			acyclic = false
			return
		} else if !(temporaryMark[u] && permanentMark[u]) {
			temporaryMark[u] = true
			for _, v := range p.jobDependency[u] {
				visit(v)
				if !acyclic {
					return
				}
			}
			delete(temporaryMark, u)
			permanentMark[u] = true
			queue = append(queue, u)
		}
	}

	for u := range p.jobDependency {
		if !(temporaryMark[u] && permanentMark[u]) {
			visit(u)
			if !acyclic {
				return nil, fmt.Errorf("%s (%d) is not a DAG", p.name, p.id)
			}
		}
	}
	return queue, nil
}

func runJob(job Job) (ExecutionStatus, error) {
	ret := ExecutionStatus{ExecutionStart: time.Now().UTC()}
	if _, err := os.Stat(job.Path); err != nil {
		ret.Code = Failed
		return ret, err
	}
	start := time.Now()
	err := exec.Command(job.Path, job.Args...).Run()
	ret.Runtime = time.Since(start)
	if err != nil {
		ret.Code = Failed
		return ret, err
	}
	ret.Code = Succeeded
	return ret, err
}

func (p pipeline) Run() (PipelineStatus, error) {
	ret := PipelineStatus{PipelineName: p.name}
	queue, err := p.topologicalSort()
	if err != nil {
		return ret, err
	} else if queue == nil {
		return ret, fmt.Errorf("%s (%d) has not been scheduled", p.name, p.id)
	}
	ret.JobStatus = make([]JobStatus, len(queue))
	var (
		executionStatus ExecutionStatus
		currentRetry    int
		nextBackoff     time.Duration
	)
	for i, jID := range queue {
		ret.JobStatus[i].Job = p.jobByID[jID]
		ret.JobStatus[i].JobID = jID
		ret.JobStatus[i].LastStatus.Code = Scheduled
	}
	ret.Status.ExecutionStart = time.Now().UTC()
	for i, jID := range queue {
		job := p.jobByID[jID]
		for {
			executionStatus, err = runJob(job)
			ret.JobStatus[i].Statuses = append(ret.JobStatus[i].Statuses, executionStatus)
			ret.JobStatus[i].LastStatus = *ret.JobStatus[i].LastExecutionStatus()
			ret.Status.Runtime += executionStatus.Runtime
			if err == nil {
				break
			}
			currentRetry++
			if job.Backoff != nil {
				nextBackoff = job.Backoff.NextBackOff()
			}
			if (currentRetry >= job.Retry && job.Retry != -1) || nextBackoff == backoff.Stop {
				ret.Status.Code = Failed
				return ret, fmt.Errorf("%s (%d) in %s (%d) returned a permanent error (%v)",
					job.Name, job.id, p.name, p.id, err)
			}
			time.Sleep(nextBackoff)
		}
	}
	ret.Status.Code = Succeeded
	return ret, nil
}
