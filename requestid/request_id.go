package requestid

import "github.com/google/uuid"

const (
	Key = "requestID"
)

type RequestIDer interface {
	GetRequestID() string
}

func New() string {
	return uuid.New().String()
}
