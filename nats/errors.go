package nats

import "fmt"

// WorkerAlreadyRunningError is an error, which represents, that worker was already started and can not be started
// again.
type WorkerAlreadyRunningError struct {
	Message string
	BaseErr error
}

func (e WorkerAlreadyRunningError) Error() string {
	template := "worker is already running."
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.Message, e.BaseErr)
	}

	return fmt.Sprintf(template, e.Message)
}

func (e WorkerAlreadyRunningError) Unwrap() error {
	return e.BaseErr
}

// WorkerAlreadyStoppedError is an error, which represents, that worker was not started yet or was already stopped.
type WorkerAlreadyStoppedError struct {
	Message string
	BaseErr error
}

func (e WorkerAlreadyStoppedError) Error() string {
	template := "worker is already stopped."
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.Message, e.BaseErr)
	}

	return fmt.Sprintf(template, e.Message)
}

func (e WorkerAlreadyStoppedError) Unwrap() error {
	return e.BaseErr
}
