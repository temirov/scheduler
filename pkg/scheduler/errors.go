package scheduler

import "errors"

var (
	ErrTaskAlreadyExists = errors.New("task with this ID already exists")
	ErrTaskNotFound      = errors.New("task not found")
)
