package jobhub

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
)

var nextPipelineID = 1

type Pipeline struct {
	name            string
	logger          logrus.FieldLogger
	id              int
	nextJobID       int
	jobContainer    []Job
	jobByID         map[int]Job
	jobDependency   map[int][]int
	startingJob     map[int]bool
	recursionLevels map[int]int
}

type Job struct {
	Name    string
	Path    string
	Args    []string
	Retry   int
	Backoff backoff.BackOff

	pipelineID int
	id         int
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

func (jStatus JobStatus) LastExecutionStatus() *ExecutionStatus {
	if len(jStatus.Statuses) > 0 {
		return &jStatus.Statuses[len(jStatus.Statuses)-1]
	}
	return nil
}

func (sCode StatusCode) String() string {
	return statusDescription[sCode]
}

func (j Job) String() string {
	return fmt.Sprintf("Pipeline ID: %d | ID: %d | Name: %s", j.pipelineID, j.id, j.Name)
}

func NewPipeline(name string, logger logrus.FieldLogger) *Pipeline {
	return &Pipeline{
		name:            name,
		logger:          logger,
		jobByID:         make(map[int]Job),
		jobDependency:   make(map[int][]int),
		startingJob:     make(map[int]bool),
		recursionLevels: make(map[int]int),
	}
}

func nextIDPipeline() int {
	tempID := nextPipelineID
	nextPipelineID++
	return tempID
}

func (p *Pipeline) nextIDJob() int {
	p.nextJobID++
	return p.nextJobID
}

func (p *Pipeline) AddJob(job Job) Job {
	for _, j := range p.jobContainer {
		if j.id == job.id {
			p.logger.Panicf("Pipeline [%d][%s] | Job [%d][%s] has already been added",
				p.id, p.name, job.id, job.Name)
		}
	}
	if p.id == 0 {
		p.id = nextIDPipeline()
	}
	job.pipelineID = p.id
	job.id = p.nextIDJob()
	p.jobContainer = append(p.jobContainer, job)
	p.jobByID[job.id] = job
	p.startingJob[job.id] = true
	return job
}

func (p *Pipeline) AddJobDependency(job Job, deps ...Job) {
	var tempContainer []Job
	tempContainer = append(deps, job)
	if p.id == 0 {
		p.logger.Panicf("Pipeline [%d][%s] | Pipeline not initialized",
			p.id, p.name)
	}
	if p.jobContainer == nil {
		p.logger.Panicf("Pipeline [%d][%s] | Job Container is empty",
			p.id, p.name)
	}
	for _, givenJob := range tempContainer {
		if givenJob.pipelineID != p.id {
			p.logger.Panicf("Pipeline [%d][%s] | Job [%d][%s] does not belong to this pipeline",
				p.id, p.name, givenJob.id, givenJob.Name)
		}
	}
	for _, d := range deps {
		p.jobDependency[job.id] = append(p.jobDependency[job.id], d.id)
		delete(p.startingJob, d.id)
	}
}

func (p *Pipeline) topologicalSort() []int {
	var (
		permMark       = map[int]bool{}
		tempMark       = map[int]bool{}
		cycle          bool
		cycleRecovery  int
		cycleRecovered []int
		queue          []int
		visit          func(int)
	)
	visit = func(u int) {
		if permMark[u] {
			return
		} else if tempMark[u] {
			cycle = true
			cycleRecovery = u
			return
		}
		tempMark[u] = true
		for _, v := range p.jobDependency[u] {
			visit(v)
			if cycle {
				if cycleRecovery > 0 {
					if cycleRecovery == u {
						cycleRecovery = 0
					}
					cycleRecovered = append(cycleRecovered, u)
				}
				return
			}
		}
		delete(tempMark, u)
		permMark[u] = true
		queue = append(queue, u)
	}
	for u := range p.jobDependency {
		if !permMark[u] {
			visit(u)
		}
		if cycle {
			p.logger.Panicf("Pipeline [%d][%s] | Job dependency graph is not a DAG (%d)",
				p.id, p.name, cycleRecovered)
		}
	}
	return queue
}

func (p *Pipeline) resolveDependencyRecursion(jobID, level int) {
	if l := p.recursionLevels[jobID]; l >= level {
		return
	}
	p.recursionLevels[jobID] = level
	level++
	for _, depID := range p.jobDependency[jobID] {
		p.resolveDependencyRecursion(depID, level)
	}
}

func (p *Pipeline) resolveDependency() []int {
	//TODO(amwolff) switch to sort.SortSlice
	var queue []int
	tempRL := make(map[int]int)
	for startID := range p.startingJob {
		p.resolveDependencyRecursion(startID, 1)
	}
	for jID, l := range p.recursionLevels {
		tempRL[jID] = l
	}
	for len(tempRL) > 0 {
		var jobID int
		max := 0
		for jID, l := range tempRL {
			if l > max {
				max = l
				jobID = jID
			}
		}
		queue = append(queue, jobID)
		delete(tempRL, jobID)
	}
	return queue
}

func (p *Pipeline) runJob(job Job) (ExecutionStatus, error) {
	ret := ExecutionStatus{ExecutionStart: time.Now().UTC()}
	_, err := os.Stat(job.Path)
	if err != nil {
		ret.Code = Failed
		return ret, err
	}
	process := exec.Command(job.Path, job.Args...)
	start := time.Now()
	err = process.Run()
	ret.Runtime = time.Since(start)
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			p.logger.Errorf("Cannot cast to exitError (%s)", err)
		}
		p.logger.Error(exitError.Sys().(syscall.WaitStatus))
		ret.Code = Failed
		return ret, err
	}
	ret.Code = Succeeded
	return ret, err
}

func (p *Pipeline) Run() PipelineStatus {
	queue := p.resolveDependency()
	if queue == nil {
		p.logger.Panicf("Pipeline [%d][%s] | No jobs in queue", p.id, p.name)
	}
	ret := PipelineStatus{
		PipelineName: p.name,
		JobStatus:    make([]JobStatus, len(queue)),
	}
	var (
		executionStatus ExecutionStatus
		err             error
		currentRetry    int
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
			executionStatus, err = p.runJob(job)
			ret.JobStatus[i].Statuses = append(ret.JobStatus[i].Statuses, executionStatus)
			ret.JobStatus[i].LastStatus = *ret.JobStatus[i].LastExecutionStatus()
			ret.Status.Runtime += executionStatus.Runtime
			if err == nil {
				break
			}
			nextBackoff := job.Backoff.NextBackOff()
			currentRetry++
			if (currentRetry >= job.Retry && job.Retry != -1) || nextBackoff == backoff.Stop {
				p.logger.Panicf("Pipeline [%d][%s] | Job [%d][%s][Runtime: %v][StatusCode: %d] | %s",
					p.id, p.name, job.id, job.Name, executionStatus.Runtime, executionStatus.Code, err)
				ret.Status.Code = Failed
				return ret
			}
			if job.Backoff != nil {
				time.Sleep(nextBackoff)
			}
		}
	}
	ret.Status.Code = Succeeded
	return ret
}
