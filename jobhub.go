package jobhub

import (
	//"github.com/cenkalti/backoff"
	//"github.com/orian/utils/common"
	"github.com/sirupsen/logrus"
	"os/exec"
	"syscall"
	"time"
)

/*
type PipelineManager interface {
	New(string) Pipeline
	AddJob(Job) Job
	AddJobDependency(...Job)
	Run()
}
*/

var nextPipelineID int = 1

type Pipeline struct {
	Name          string
	Log           logrus.FieldLogger
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

type exitStatus struct {
	runtime time.Duration
	status  syscall.WaitStatus
}

func (p *Pipeline) runJob(job Job) (*exitStatus, error) {
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
			p.Log.Panicf("Cannot cast to exitError: %s", err)
		}
		return &exitStatus{runtime: elapsed, status: exitError.Sys().(syscall.WaitStatus)}, err
	}
	return nil, err
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
			/* error */
		}
	}
	if p.id == 0 {
		p.id = nextIDPipeline()
		/*
			if p.Log == nil {
			 how do you log an error without a logger?
			}
		*/
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
func (p *Pipeline) getJobIndex(job Job) int {
	for pos, val := range p.jobContainer {
		if val == job {
			return pos
		}
	}
	return -1
}
*/
