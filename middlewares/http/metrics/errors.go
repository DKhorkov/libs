package metrics

import "errors"

// ErrHijackUnsupported — обёрнутый ResponseWriter не умеет Hijack.
var ErrHijackUnsupported = errors.New("metrics: response writer does not implement http.Hijacker")
