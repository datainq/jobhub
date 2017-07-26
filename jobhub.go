package jobhub

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

var nextPipelineID = 1

type Pipeline struct {
	Name string
	Log  logrus.FieldLogger

	id              int
	nextJobID       int
	jobContainer    []Job
	jobByID         map[int]Job
	jobDependency   map[int][]int
	startingJob     map[int]bool
	recursionLevels map[int]int
}

type Job struct {
	Name string
	Path string

	pipelineID int
	id         int
}

type PipelineStatus struct {
	Code      StatusCode
	Runtime   time.Duration
	JobStatus []JobStatus
}

type JobStatus struct {
	Job      Job
	Code     StatusCode
	Statuses []ExecutionStatus
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

func (j Job) String() string {
	return fmt.Sprintf("Pipeline ID: %d | ID: %d| Name: %s", j.pipelineID, j.id, j.Name)
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		Log:             logrus.StandardLogger(),
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
			p.Log.Panicf("Pipeline [%d][%s] | Job [%d][%s] | Job has already been added",
				p.id, p.Name, job.id, job.Name)
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
		p.Log.Panicf("Pipeline [%d][%s] | Pipeline not initialized",
			p.id, p.Name)
	}
	if p.jobContainer == nil {
		p.Log.Panicf("Pipeline [%d][%s] | Job Container is empty",
			p.id, p.Name)
	}
	for _, givenJob := range tempContainer {
		if givenJob.pipelineID != p.id {
			p.Log.Panicf("Pipeline [%d][%s] | Job [%d][%s] does not belong to this pipeline",
				p.id, p.Name, givenJob.id, givenJob.Name)
		}
	}
	for _, d := range deps {
		p.jobDependency[job.id] = append(p.jobDependency[job.id], d.id)
		delete(p.startingJob, d.id)
	}
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
	for startID, _ := range p.startingJob {
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
	process := exec.Command(job.Path)
	start := time.Now()
	err = process.Run()
	ret.Runtime = time.Since(start)
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			p.Log.Errorf("Cannot cast to exitError: %s", err)
		}
		p.Log.Debug(exitError.Sys().(syscall.WaitStatus))
		ret.Code = Failed
		return ret, err
	}
	ret.Code = Succeeded
	return ret, err
}

func (p *Pipeline) Run() PipelineStatus {
	queue := p.resolveDependency()
	if queue == nil {
		p.Log.Panicf("Pipeline [%d][%s] | No jobs in queue", p.id, p.Name)
	}
	pipelineStatus := PipelineStatus{JobStatus: make([]JobStatus, len(queue))}
	pipelineStatus.Code = Scheduled
	for i, jID := range queue {
		pipelineStatus.JobStatus[i].Job = p.jobByID[jID]
		pipelineStatus.JobStatus[i].Code = Scheduled
	}
	for i, jID := range queue {
		executionStatus, err := p.runJob(p.jobByID[jID])
		pipelineStatus.JobStatus[i].Statuses = append(pipelineStatus.JobStatus[i].Statuses, executionStatus)
		pipelineStatus.JobStatus[i].Code = executionStatus.Code
		pipelineStatus.Runtime += executionStatus.Runtime
		if err != nil {
			p.Log.Errorf("Pipeline [%d][%s] | Job [%d][%s][Runtime: %v][StatusCode: %d] | %s",
				p.id, p.Name, jID, p.jobByID[jID].Name, executionStatus.Runtime, executionStatus.Code, err)
			pipelineStatus.Code = Failed
			return pipelineStatus
		}
	}
	pipelineStatus.Code = Succeeded
	return pipelineStatus
}

//TODO(amwolff) remove this method
func (p *Pipeline) PrintDeps() {
	p.Log.Debugf("Queue: %d", p.resolveDependency())
}
