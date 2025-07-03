package nats

import "fmt"

// ConsumerAlreadyRunningError is an error, which represents, that consumer was already started and can not be started
// again.
type ConsumerAlreadyRunningError struct {
	Message string
	BaseErr error
}

func (e ConsumerAlreadyRunningError) Error() string {
	template := "consumer is already running"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e ConsumerAlreadyRunningError) Unwrap() error {
	return e.BaseErr
}

// ConsumerAlreadyStoppedError is an error, which represents, that consumer was not started yet or was already stopped.
type ConsumerAlreadyStoppedError struct {
	Message string
	BaseErr error
}

func (e ConsumerAlreadyStoppedError) Error() string {
	template := "consumer is already stopped"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e ConsumerAlreadyStoppedError) Unwrap() error {
	return e.BaseErr
}
