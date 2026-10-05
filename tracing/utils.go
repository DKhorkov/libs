package tracing

import (
	"fmt"
	"runtime"

	"go.opentelemetry.io/otel/trace"
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

// RecordError отмечает ошибку на спане: событие exception с текстом и статус
// Error. Пустая ошибка спан не трогает, поэтому декоратор зовёт хелпер на любом
// исходе, не проверяя его сам.
//
// Без этого спан слоя оставался зелёным при любом исходе: ошибка проходила
// сквозь декоратор наверх, и по трассе не было видно, на каком слое случился
// сбой.
func RecordError(span trace.Span, err error) {
	if err == nil {
		return
	}

	span.RecordError(err)
	span.SetStatus(StatusError, err.Error())
}
