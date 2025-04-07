package job_test

import (
	"fmt"
	"testing"

	"github.com/golamee/job/job"
)

type Man struct {
	Name string
	Age  int
}

var j job.Job[Man]

func TestCreateJobTest(t *testing.T) {
	j = job.New(func(value Man) {
		fmt.Println(value)
	})

}

func TestDispatchJobSuccess(t *testing.T) {
	j.Dispatch(Man{Name: "TestDispatchJobSuccess", Age: 30})
}

func TestDispatchJobFailed(t *testing.T) {
	j.Dispatch(Man{Name: "TestDispatchJobFailed", Age: 30})
}
