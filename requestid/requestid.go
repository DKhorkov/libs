package requestid

import "github.com/google/uuid"

const (
	Key = "requestID"
)

func New() string {
	return uuid.New().String()
}
