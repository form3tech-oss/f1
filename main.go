package main

import (
	"fmt"
	"time"

	"github.com/form3tech-oss/f1/pkg/f1"

	"github.com/form3tech-oss/f1/pkg/f1/testing"
)

func main() {
	testing.Add("template", setupMySuperFastLoadTest)
	f1.Execute()
}

func setupMySuperFastLoadTest(t *testing.T) (testing.RunFn, testing.TeardownFn) {
	runFn := func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
		//fmt.Println("Do some work")
		//fmt.Printf("Env received: %s\n", os.Getenv("SOP"))
		//fmt.Printf("Env received: %s\n", os.Getenv("SIP"))
		//fmt.Printf("Env received: %s\n", os.Getenv("FDP"))
		//require.Fail(t, "dummy failure")
	}

	teardownFn := func(t *testing.T) {
		fmt.Println("Teardown")
	}

	return runFn, teardownFn
}
