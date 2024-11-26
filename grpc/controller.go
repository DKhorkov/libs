package grpc

import (
	"log/slog"

	"google.golang.org/grpc"
)

// Controller is gRPC controller, which will be used in app.
type Controller struct {
	Server *grpc.Server
	Host   string
	Port   int
	Logger *slog.Logger
}
