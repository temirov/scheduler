package main

import (
	"github.com/temirov/scheduler/pkg/scheduler"
)

func main() {
	// This calls the CLI code from the scheduler library.
	// Because we import screener/tasks (with init() funcs),
	// the tasks are registered on the global scheduler.
	scheduler.Execute()
}
