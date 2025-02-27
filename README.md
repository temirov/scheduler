# Scheduler

A simple Go scheduler with minimal dependencies.

## Overview

The Scheduler is a lightweight task scheduling library written in Go. It allows you to register tasks with various scheduling options, such as daily, interval-based, or one-time execution. The scheduler is designed to be easy to integrate into existing Go projects and provides a simple CLI for managing tasks.

## Features

- **Task Registration**: Register tasks with unique identifiers and descriptions.
- **Flexible Scheduling**: Supports daily, interval, and one-time schedules.
- **Error Handling**: Provides error feedback for task registration and execution.
- **CLI Integration**: Manage tasks via command-line interface with options to list, run, and start tasks.
- **Concurrency**: Executes tasks concurrently and handles retries on failure.

## Installation

To use the Scheduler in your project, you can import it as a module:

```bash
go get github.com/temirov/scheduler
```

## Usage

### Registering a Task

To register a task, you need to define a task that implements the `Task` interface and register it with a schedule.

```go
package main

import (
  "github.com/temirov/scheduler/pkg/scheduler"
)

func main() {
    // Define a daily schedule
    dailySchedule := scheduler.DailySchedule{Hour: 9, Minute: 0}
    // Create a new task
    myTask := NewMyTask("my-task-id", dailySchedule)
    // Register the task
    err := scheduler.RegisterTaskInfo("my-task-id", "My daily task", dailySchedule)
    if err != nil {
        panic(err)
    }
}
```

### Running Tasks

To run a task, you can use the `RunTask` function:

```go
schedulerInstance := scheduler.NewScheduler()
schedulerInstance.Start()
```

### CLI Commands

The Scheduler provides a command-line interface for managing tasks. Here are some of the available commands:

- **List Tasks**: `scheduler --list`
- **Run a Task Immediately**: `scheduler --run <task_id>`
- **Start the Scheduler**: `scheduler --start`

### Example

Here's a complete example of registering and running a task:

```go
package main
import (
    "github.com/temirov/scheduler/pkg/scheduler"
    "time"
)
func main() {
    // Define a one-time schedule
    oneTimeSchedule := scheduler.NewOneTimeSchedule(time.Now().Add(1 time.Hour))
    // Create and register a task
    taskID := "one-time-task"
    err := scheduler.RegisterTaskInfo(taskID, "A one-time task", oneTimeSchedule)
    if err != nil {
        panic(err)
    }
    // Start the scheduler
    schedulerInstance := scheduler.NewScheduler()
    schedulerInstance.Start()
    // Run the task immediately
    err = schedulerInstance.RunTaskNow(taskID)
    if err != nil {
        panic(err)
    }
}
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## License

This project is licensed under the MIT License.
