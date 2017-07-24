package jobhub

import "testing"

func TestAddJobDependency(t *testing.T) {
	var want error
	p := NewPipeline()
	jA := Job{
		Name: "A",
		Path: "./tests/simple_success",
	}
	p.AddJobDependency(jA)
	if want == nil {
		t.Error("No error detected and there should be one.")
	}
}
