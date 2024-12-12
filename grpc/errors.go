package grpc

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BaseError is a base gRPC error, which is used to inform clients, that something went wrong between microservices.
type BaseError struct {
	Message string     `json:"message"`
	BaseErr error      `json:"baseErr"`
	Status  codes.Code `json:"-"`
}

func (e BaseError) Error() string {
	if e.BaseErr != nil {
		return fmt.Sprintf("%s. Base error: %v", e.Message, e.BaseErr)
	}

	return e.Message
}

// GRPCStatus is a member function, which is used by gRPC when converting an error into a status.
func (e BaseError) GRPCStatus() *status.Status {
	return status.New(e.Status, e.Error())
}
