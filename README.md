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
    
    // Register the task with a single call
    err := scheduler.RegisterTask(
        "my-task-id",
        "My daily task", 
        dailySchedule,
        func() scheduler.Task {
            return NewMyTask()
        },
    )
    if err != nil {
        panic(err)
    }
}
```

Using init() function for registration:

```go
func init() {
    // Define a daily schedule
    dailySchedule := scheduler.DailySchedule{Hour: 9, Minute: 0}
    
    // Register task with a single call
    scheduler.RegisterTask(
        "my-task-id",
        "My daily task", 
        dailySchedule,
        func() scheduler.Task {
            return NewMyTask()
        },
    )
}

// NewMyTask creates an instance of your task
func NewMyTask() scheduler.Task {
    return &MyTask{}
}
```

> **Note:** For backward compatibility, the separate `RegisterTaskInfo` and `RegisterTaskFactory` functions are still available, but using the combined `RegisterTask` function is recommended.

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
    "context"
    "github.com/temirov/scheduler/pkg/scheduler"
    "time"
)

// Example task implementation
type ExampleTask struct{}

func (t *ExampleTask) ID() string {
    return "one-time-task"
}

func (t *ExampleTask) Schedule() scheduler.TimeSchedule {
    return scheduler.NewOneTimeSchedule(time.Now().Add(1 * time.Hour))
}

func (t *ExampleTask) MaxRetries() int {
    return 3
}

func (t *ExampleTask) RetryDelay(attempt int) time.Duration {
    return time.Second * time.Duration(attempt*5)
}

func (t *ExampleTask) BeforeExecute(ctx context.Context) error {
    return nil
}

func (t *ExampleTask) Run(ctx context.Context) error {
    // Task implementation
    return nil
}

func NewExampleTask() scheduler.Task {
    return &ExampleTask{}
}

func main() {
    // Define a one-time schedule
    oneTimeSchedule := scheduler.NewOneTimeSchedule(time.Now().Add(1 * time.Hour))
    
    taskID := "one-time-task"
    
    // Register task with a single call
    err := scheduler.RegisterTask(
        taskID,
        "A one-time task",
        oneTimeSchedule,
        func() scheduler.Task {
            return NewExampleTask()
        },
    )
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
