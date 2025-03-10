package scheduler

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/temirov/scheduler/pkg/utils"
)

// Execute processes CLI flags and executes the requested command.
// This function is intended to be called from the destination project's main package.
func Execute() {
	listCommand := flag.Bool("list", false, "List all registered tasks")
	runCommand := flag.String("run", "", "Run a specific task immediately")
	startCommand := flag.Bool("start", false, "Start the scheduler with all registered tasks")
	helpCommand := flag.Bool("help", false, "Show detailed help for tasks")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	// If no flags are provided, default to listing tasks.
	if !*listCommand && !*helpCommand && *runCommand == "" && !*startCommand {
		*listCommand = true
	}
	if *listCommand {
		listTasks()
		return
	}
	if *helpCommand {
		showHelp()
		return
	}
	if *runCommand != "" {
		runTaskFromRegistry(*runCommand)
		return
	}
	if *startCommand {
		startScheduler()
		return
	}
}

func listTasks() {
	taskInfos := GetAllTaskInfo()
	if len(taskInfos) == 0 {
		fmt.Println("No tasks are currently registered in the scheduler.")
		fmt.Println("Make sure task modules are properly imported.")
		return
	}
	currentTime := time.Now()
	var validTaskInfos []TaskInfo
	var invalidTaskIDs []string
	for _, taskInfo := range taskInfos {
		if taskInfo.Schedule == nil {
			invalidTaskIDs = append(invalidTaskIDs, taskInfo.ID)
			continue
		}
		nextRunPtr := taskInfo.Schedule.NextRun(currentTime)
		if nextRunPtr == nil || !nextRunPtr.After(currentTime) {
			invalidTaskIDs = append(invalidTaskIDs, taskInfo.ID)
			continue
		}
		validTaskInfos = append(validTaskInfos, taskInfo)
	}
	if len(validTaskInfos) == 0 {
		fmt.Println("Error: No tasks with valid schedules are registered.")
		if len(invalidTaskIDs) > 0 {
			fmt.Println("The following tasks have invalid schedules (no future run time):")
			for _, taskID := range invalidTaskIDs {
				fmt.Printf("  - %s\n", taskID)
			}
		}
		return
	}
	if len(invalidTaskIDs) > 0 {
		fmt.Println("Warning: Some tasks have invalid schedules (no future run time):")
		for _, taskID := range invalidTaskIDs {
			fmt.Printf("  - %s\n", taskID)
		}
		fmt.Println("")
	}
	fmt.Println("Scheduler Tasks")
	fmt.Println("==============")
	fmt.Println("")
	taskIDWidth := 25
	scheduleWidth := 40
	nextRunWidth := 25
	fmt.Printf("%-*s %-*s %-*s\n", taskIDWidth, "TASK ID", scheduleWidth, "SCHEDULE", nextRunWidth, "NEXT RUN")
	separator := ""
	totalWidth := taskIDWidth + scheduleWidth + nextRunWidth + 2
	for index := 0; index < totalWidth; index++ {
		separator += "-"
	}
	fmt.Println(separator)
	for _, taskInfo := range validTaskInfos {
		taskID := taskInfo.ID
		scheduleDesc := taskInfo.Schedule.Description()
		nextRunPtr := taskInfo.Schedule.NextRun(currentTime)
		nextRunDesc := formatNextRunTime(nextRunPtr)
		scheduleLines := utils.WordWrap(scheduleDesc, scheduleWidth)
		fmt.Printf("%-*s %-*s %-*s\n", taskIDWidth, taskID, scheduleWidth, scheduleLines[0], nextRunWidth, nextRunDesc)
		for subIndex := 1; subIndex < len(scheduleLines); subIndex++ {
			fmt.Printf("%-*s %-*s %-*s\n", taskIDWidth, "", scheduleWidth, scheduleLines[subIndex], nextRunWidth, "")
		}
	}
	fmt.Println("\n(Run with --help for more information)")
}

func runTaskFromRegistry(taskID string) {
	taskInfo, err := GetTaskInfo(taskID)
	if err != nil {
		fmt.Printf("Error: Task '%s' is not registered: %v\n", taskID, err)
		fmt.Println("Run with --list to see available tasks.")
		os.Exit(1)
	}
	taskInstance, err := createTaskFromInfo(taskInfo)
	if err != nil {
		fmt.Printf("Error creating task '%s': %v\n", taskID, err)
		os.Exit(1)
	}
	ctx, cancelFunction := context.WithCancel(context.Background())
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChannel
		fmt.Println("\nReceived interrupt signal, canceling task...")
		cancelFunction()
	}()
	fmt.Printf("Running task '%s'...\n", taskID)
	if err := taskInstance.BeforeExecute(ctx); err != nil {
		fmt.Printf("Task '%s' failed in BeforeExecute: %v\n", taskID, err)
		os.Exit(1)
	}
	startTime := time.Now()
	if err := taskInstance.Run(ctx); err != nil {
		fmt.Printf("Task '%s' failed after %v: %v\n", taskID, time.Since(startTime), err)
		os.Exit(1)
	}
	fmt.Printf("Task '%s' completed successfully in %v.\n", taskID, time.Since(startTime))
}

func startScheduler() {
	taskInfos := GetAllTaskInfo()
	if len(taskInfos) == 0 {
		fmt.Println("No tasks are registered in the scheduler.")
		fmt.Println("Make sure task modules are properly imported.")
		return
	}
	currentTime := time.Now()
	var validTaskInfos []TaskInfo
	var invalidTaskIDs []string
	for _, taskInfo := range taskInfos {
		if taskInfo.Schedule == nil {
			invalidTaskIDs = append(invalidTaskIDs, taskInfo.ID)
			continue
		}
		nextRunPtr := taskInfo.Schedule.NextRun(currentTime)
		if nextRunPtr == nil || !nextRunPtr.After(currentTime) {
			invalidTaskIDs = append(invalidTaskIDs, taskInfo.ID)
			continue
		}
		validTaskInfos = append(validTaskInfos, taskInfo)
	}
	if len(validTaskInfos) == 0 {
		fmt.Println("Error: Cannot start scheduler - no tasks with valid schedules.")
		if len(invalidTaskIDs) > 0 {
			fmt.Println("The following tasks have invalid schedules (no future run time):")
			for _, taskID := range invalidTaskIDs {
				fmt.Printf("  - %s\n", taskID)
			}
		}
		return
	}
	if len(invalidTaskIDs) > 0 {
		fmt.Println("Warning: Some tasks have invalid schedules and will not be scheduled:")
		for _, taskID := range invalidTaskIDs {
			fmt.Printf("  - %s\n", taskID)
		}
		fmt.Println("")
	}
	schedulerInstance := NewScheduler()
	var registeredTaskIDs []string
	var failedTaskIDs []string
	for _, taskInfo := range validTaskInfos {
		taskInstance, err := createTaskFromInfo(taskInfo)
		if err != nil {
			failedTaskIDs = append(failedTaskIDs, taskInfo.ID)
			slog.Error("Failed to create task", "task_id", taskInfo.ID, "error", err)
			continue
		}
		if err := schedulerInstance.RegisterTask(taskInstance); err != nil {
			failedTaskIDs = append(failedTaskIDs, taskInfo.ID)
			slog.Error("Failed to register task", "task_id", taskInfo.ID, "error", err)
			continue
		}
		registeredTaskIDs = append(registeredTaskIDs, taskInfo.ID)
	}
	if len(registeredTaskIDs) == 0 {
		fmt.Println("Error: No tasks could be registered with the scheduler.")
		return
	}
	if len(failedTaskIDs) > 0 {
		fmt.Println("Warning: Some tasks could not be registered:")
		for _, taskID := range failedTaskIDs {
			fmt.Printf("  - %s\n", taskID)
		}
		fmt.Println("")
	}
	fmt.Printf("Starting scheduler with %d tasks...\n", len(registeredTaskIDs))
	fmt.Println("Registered tasks:")
	for _, taskID := range registeredTaskIDs {
		taskInfo, err := GetTaskInfo(taskID)
		if err != nil {
			fmt.Printf("Warning: Could not get info for task '%s': %v\n", taskID, err)
			continue
		}
		nextRunPtr := taskInfo.Schedule.NextRun(currentTime)
		fmt.Printf("  - %s (next run: %s)\n", taskID, formatNextRunTime(nextRunPtr))
	}
	schedulerInstance.Start()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Scheduler running. Press Ctrl+C to stop.")
	<-sigChan
	fmt.Println("\nShutting down scheduler...")
	schedulerInstance.Stop()
	fmt.Println("Scheduler stopped")
}

func createTaskFromInfo(taskInfo TaskInfo) (Task, error) {
	factory, exists := GetTaskFactory(taskInfo.ID)
	if !exists {
		return nil, fmt.Errorf("no factory registered for task: %s", taskInfo.ID)
	}

	return factory(taskInfo)
}

func formatNextRunTime(nextRunPtr *time.Time) string {
	if nextRunPtr == nil {
		return "Unknown"
	}
	nextRun := *nextRunPtr
	currentTime := time.Now()
	if utils.IsSameDay(nextRun, currentTime) {
		return fmt.Sprintf("Today at %02d:%02d", nextRun.Hour(), nextRun.Minute())
	}
	tomorrow := currentTime.Add(24 * time.Hour)
	if utils.IsSameDay(nextRun, tomorrow) {
		return fmt.Sprintf("Tomorrow at %02d:%02d", nextRun.Hour(), nextRun.Minute())
	}
	return nextRun.Format("Mon, Jan 2 at 15:04")
}

func showHelp() {
	fmt.Println("Scheduler CLI Help")
	fmt.Println("------------------")
	fmt.Println("--list              List all registered tasks with their schedules")
	fmt.Println("--run <task_id>     Run a specific task immediately")
	fmt.Println("--start             Start the scheduler with all registered tasks")
	fmt.Println("--help              Show this help message")
	fmt.Println("")
	fmt.Println("Available Tasks:")
	taskInfos := GetAllTaskInfo()
	currentTime := time.Now()
	for _, taskInfo := range taskInfos {
		scheduleStatus := "Valid schedule"
		if taskInfo.Schedule == nil {
			scheduleStatus = "No schedule defined"
		} else {
			nextRunPtr := taskInfo.Schedule.NextRun(currentTime)
			if nextRunPtr == nil || !nextRunPtr.After(currentTime) {
				scheduleStatus = "Invalid schedule (no future run time)"
			}
		}
		fmt.Printf("  %-20s %s\n", taskInfo.ID, taskInfo.Description)
		fmt.Printf("    %s\n", scheduleStatus)
	}
}
