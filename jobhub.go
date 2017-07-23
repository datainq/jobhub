package jobhub

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

var nextPipelineID int = 1

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

type exitStatus struct {
	runtime time.Duration
	status  syscall.WaitStatus
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
			p.Log.Panicf("Pipeline [%d][%s] | Job [%d][%s] | Panic: Job has already been added",
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
	for _, d := range deps {
		p.jobDependency[job.id] = append(p.jobDependency[job.id], d.id)
		delete(p.startingJob, d.id)
	}
}

func (p *Pipeline) resolveDependencyRecursion(jobID int, level int) {
	if l := p.recursionLevels[jobID]; l >= level {
		return
	}
	p.recursionLevels[jobID] = level
	level++
	for _, depID := range p.jobDependency[jobID] {
		p.Log.Debugf("%s -> %s | recursion level: %d", p.jobByID[jobID], p.jobByID[depID], level)
		p.resolveDependencyRecursion(depID, level)
	}
}

func (p *Pipeline) resolveDependency() []int {
	var queue []int
	tempRL := make(map[int]int)
	for startID, _ := range p.startingJob {
		p.resolveDependencyRecursion(startID, 1)
	}
	for jID, l := range p.recursionLevels {
		tempRL[jID] = l
	}
	for i := 0; i < len(tempRL); {
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

func (p *Pipeline) runJob(job Job) (*exitStatus, error) {
	_, err := os.Stat(job.Path)
	if err != nil {
		return nil, err
	}
	process := exec.Command(job.Path)
	start := time.Now()
	err = process.Run()
	elapsed := time.Since(start)
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		p.Log.Panicf("Cannot cast to exitError", err)
	}
	return &exitStatus{runtime: elapsed, status: exitError.Sys().(syscall.WaitStatus)}, err
}

func (p *Pipeline) Run() {
	if p.jobDependency != nil {
		queue := p.resolveDependency()
		for _, jID := range queue {
			exitStatus, err := p.runJob(p.jobByID[jID])
			if err != nil {
				p.Log.Panicf("Pipeline [%d][%s] | Job [%d][%s][t: %f] | Panic: %s",
					p.id, p.Name, jID, p.jobByID[jID].Name, exitStatus.runtime, err)
			}
		}
	} else {
		p.Log.Panicf("Pipeline [%d][%s] | Panic: No jobs in queue", p.id, p.Name)
	}
}

func (p *Pipeline) PrintDeps() {
	p.Log.Debugf("Queue: %d", p.resolveDependency())
}

func (j Job) String() string {
	return fmt.Sprintf("ID (Name): %d (%s)", j.id, j.Name)
}
