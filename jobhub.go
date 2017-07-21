package jobhub

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
	"time"
)

var nextPipelineID int = 1

type Pipeline struct {
	Name            string
	Log             logrus.FieldLogger
	id, nextJobID   int
	jobContainer    []Job
	jobByID         map[int]Job
	jobDependency   map[int][]int
	isStartingJob   map[int]bool
	recursionLevels map[int]int
}

type Job struct {
	Name, Path     string
	pipelineID, id int
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
		isStartingJob:   make(map[int]bool),
		recursionLevels: make(map[int]int),
	}
}

func nextIDPipeline() int {
	tempID := nextPipelineID
	nextPipelineID++
	return tempID
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
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			p.Log.Panicf("Pipeline [%d][%s] | Job [%d][%s][t: %f] | Panic: %s", p.id, p.Name, job.id, job.Name, elapsed, err)
		}
		return &exitStatus{runtime: elapsed, status: exitError.Sys().(syscall.WaitStatus)}, err
	}
	return nil, err
}

func (p *Pipeline) nextIDJob() int {
	p.nextJobID++
	return p.nextJobID
}

func (p *Pipeline) AddJob(job Job) Job {
	for _, j := range p.jobContainer {
		if j.id == job.id {
			p.Log.Panicf("Pipeline [%d][%s] | Job [%d][%s] | Panic: Job has already been added", p.id, p.Name, job.id, job.Name)
		}
	}
	if p.id == 0 {
		p.id = nextIDPipeline()
	}
	job.pipelineID = p.id
	job.id = p.nextIDJob()
	p.jobContainer = append(p.jobContainer, job)
	p.jobByID[job.id] = job
	p.isStartingJob[job.id] = true
	return job
}

func (p *Pipeline) AddJobDependency(job Job, deps ...Job) {
	for _, d := range deps {
		p.jobDependency[job.id] = append(p.jobDependency[job.id], d.id)
		delete(p.isStartingJob, d.id)
	}
}

func (p *Pipeline) resolveDependeciesRecursion(jobID int, level int) {
	if lvl := p.recursionLevels[jobID]; lvl >= level {
		return
	}
	p.recursionLevels[jobID] = level
	level++
	for _, depID := range p.jobDependency[jobID] {
		fmt.Printf("%s -> %s | recursion level: %d\n", p.jobByID[jobID], p.jobByID[depID], level)
		p.resolveDependeciesRecursion(depID, level)
	}
}

func (p *Pipeline) resolveDependencies() {
	for start, _ := range p.isStartingJob {
		p.resolveDependeciesRecursion(p.jobContainer[start].id, 1)
	}
}

//debug func
func (p *Pipeline) PrintDeps() {
	p.resolveDependeciesRecursion(1, 1)
	fmt.Println(p.recursionLevels)
}

func (j Job) String() string {
	return fmt.Sprintf("N: %s ID: %d", j.Name, j.id)
}

/* this won't compile for a while
func (p *Pipeline) Run() {
	if p.jobDependency != nil {
		for _, job := range p.jobDependency {
			exitStatus, err := p.runJob(job)
			if err != nil {
				p.Log.Panicf("Pipeline [%d][%s] | Job [%d][%s][t: %f] | Panic: %s", p.id, p.Name, job.id, job.Name, exitStatus.runtime, err)
			}
		}
	} else {
		p.Log.Panicf("Pipeline [%d][%s] | Panic: No jobs in queue", p.id, p.Name)
	}
}
*/
