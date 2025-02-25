package logging

import (
	"context"
)

// Logger interface is created for usage in external application according to
// "dependency inversion principle" of SOLID due to working via abstractions.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/logger.go -package=mocks
type Logger interface {
	Debug(msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
	Info(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	Warn(msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	Error(msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}
