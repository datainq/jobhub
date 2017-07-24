package jobhub

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestAddJobDependency(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic.")
		}
	}()
	p := NewPipeline()
	logrus.SetLevel(logrus.DebugLevel)
	jA := Job{
		Name: "A",
		Path: "./tests/simple_success",
	}
	jB := p.AddJob(Job{
		Name: "B",
		Path: "./tests/simple_success",
	})
	p.AddJobDependency(jA, jB)
}
