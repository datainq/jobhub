package jobhub

import (
	//"github.com/cenkalti/backoff"
	//"github.com/orian/utils/common"
	"github.com/sirupsen/logrus"
	//"math/rand"
	"os/exec"
	"syscall"
	"time"
)

var log = logrus.New()

/*
type PipelineManager interface {
	New(string) Pipeline
	AddJob(Job) Job
	AddJobDependency(...Job)
	Run()
}
*/

var nextPipelineID int

type Pipeline struct {
	Name          string
	id            int
	nextJobID     int
	jobContainer  []Job
	jobDependency []Job
}

type Job struct {
	Name       string
	Path       string
	pipelineID int
	id         int
}

func (p *Pipeline) AddJob(job Job) Job {
	for _, j := range p.jobContainer {
		if j.id == job.id {
			/* error */
		}
	}
	if nextPipelineID != p.id {
		p.id = nextIDPipeline()
	}
	job.pipelineID = p.id
	job.id = p.nextIDJob()
	p.jobContainer = append(p.jobContainer, job)
	return job
}

func (p *Pipeline) AddJobDependency(jobs ...Job) {
	p.jobDependency = jobs
}

func (p *Pipeline) Run() {
	if p.jobDependency != nil {
		for _, job := range p.jobDependency {
			_, err := runJob(job)
			if err != nil {
				/* abort, I think?
				   throw an error: job[i].Name with ID: job[i].id failed */
			}
		}
	} else {
		/* error */
	}
}

/*
func (p *Pipeline) getJobIndex(j Job) int {
	for pos, val := range p.jobContainer {
		if val == j {
			return pos
		}
	}
	return -1
}
*/

func nextIDPipeline() int {
	tempID := nextPipelineID
	nextPipelineID++
	return tempID
}

func (p *Pipeline) nextIDJob() int {
	p.nextJobID++
	return p.nextJobID
}

type exitStatus struct {
	runtime time.Duration
	status  syscall.WaitStatus
}

func runJob(job Job) (*exitStatus, error) {
	_, err := exec.LookPath(job.Path)
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
			log.Panicf("Cannot cast to exitError: %s", err)
		}
		return &exitStatus{runtime: elapsed, status: exitError.Sys().(syscall.WaitStatus)}, err
	}
	return nil, err
}
