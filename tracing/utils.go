package tracing

import (
	"fmt"
	"runtime"
)

const (
	DefaultSkipLevel = 1
)

// CallerName return info about function, where trace.Span was created
// https://stackoverflow.com/questions/25927660/how-to-get-the-current-function-name
func CallerName(skipLevel int) string {
	pc, file, line, ok := runtime.Caller(skipLevel)
	if !ok {
		return fmt.Sprintf("%s on line %d: %s", "Unknown", 0, "Unknown")
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return fmt.Sprintf("%s on line %d: %s", file, line, "Unknown")
	}

	return fn.Name()
}
