package main

import (
	//"fmt"
	//"github.com/cenkalti/backoff"
	//"github.com/orian/utils/common"
	"github.com/sirupsen/logrus"
	"math/rand"
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

type Pipeline struct {
	name          string
	id            int
	jobContainer  []Job
	jobDependency []Job
}

type Job struct {
	Name       string
	Path       string
	pipelineId *int
	id         int
}

func (p Pipeline) New(name string) Pipeline {
	p.id = getRandomId()
	p.name = name
	return p
}

func (p *Pipeline) AddJob(job Job) Job {
	for _, j := range p.jobContainer {
		if j.id == job.id {
			/* error */
		}
	}
	job.pipelineId = &p.id
	job.id = getRandomId()
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

func getRandomId() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int()
}

type exitStatus struct {
	runtime time.Duration
	status  syscall.WaitStatus
}

func runJob(job Job) (*exitStatus, error) {
	var (
		start   time.Time
		elapsed time.Duration
	)
	_, err := exec.LookPath(job.Path)
	if err != nil {
		return nil, err
	}
	process := exec.Command(job.Path)
	start = time.Now()
	err = process.Run()
	elapsed = time.Since(start)
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			log.Panicf("Cannot cast to exitError: %s", err)
		}
		return &exitStatus{runtime: elapsed, status: exitError.Sys().(syscall.WaitStatus)}, err
	}
	return nil, err
}

func main() {
	/* For testing purposes only:
	//common.InitLogrus("pisch", "DEBUG", "/tmp/pisch", false, false)
	jobs := []Job{
		Job{"success", "./simple_success"}, Job{"success2", "./simple_success"}, Job{"success3", "./simple_success"}, Job{"failure", "./simple_failure"},
	}
	for _, v := range jobs {
		_, err := runJob(v)
		if err != nil {
			fmt.Println(err)
		}
	}
	*/
}
