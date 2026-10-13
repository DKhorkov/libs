package logging

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/DKhorkov/libs/logging"
)

// Option — функциональная опция New.
type Option func(*options)

// options — настройки middleware логирования.
type options struct {
	// sensitiveFields — поля верхнего уровня тела запроса, которые не
	// пишутся в лог.
	sensitiveFields []string

	// requestURL — адрес запроса для лога. nil — пишется r.URL целиком и
	// отдельным полем Query.
	requestURL func(*http.Request) string

	// level — уровень строк запроса и ответа: INFO или DEBUG.
	level logging.Level
}

// newOptions — настройки по умолчанию с применёнными опциями.
func newOptions(opts ...Option) *options {
	//nolint:exhaustruct // остальное пусто: опций нет — поведение прежнее
	result := &options{level: logging.Levels.INFO}

	for _, opt := range opts {
		opt(result)
	}

	return result
}

// WithSensitiveFields — поля верхнего уровня тела запроса (JSON), которые не
// пишутся в лог: пароли, почта.
func WithSensitiveFields(fields ...string) Option {
	return func(o *options) {
		o.sensitiveFields = append(o.sensitiveFields, fields...)
	}
}

// WithRequestURL — адрес запроса для лога вместо r.URL целиком.
//
// В пути и query бывают секреты: токен сброса пароля в пути — смена чужого
// пароля для любого, кто читает журнал, пока токен жив; почта в query —
// персональные данные. Middleware о них не знает, знает приложение: у него
// есть шаблон маршрута. Строка из функции пишется полем URL в обе строки лога,
// отдельного поля Query тогда нет — query в логе ровно то, что вернула
// функция.
func WithRequestURL(requestURL func(*http.Request) string) Option {
	return func(o *options) {
		o.requestURL = requestURL
	}
}

// WithLevel — уровень строк запроса и ответа. По умолчанию INFO; DEBUG и ниже —
// DEBUG, остальное — INFO: строка запроса не предупреждение и не ошибка.
//
// Две строки на каждый запрос — почти весь объём журнала, и приложению, которое
// держит уровень INFO ради редких событий (старт и остановка серверов), они
// нужны только при отладке.
func WithLevel(level logging.Level) Option {
	return func(o *options) {
		o.level = level
	}
}

// logLine — функция записи строки запроса и ответа на уровне options.level.
func (o *options) logLine() func(ctx context.Context, logger logging.Logger, msg string, args ...any) {
	if o.level <= logging.Levels.DEBUG {
		return logging.LogDebugContext
	}

	return logging.LogInfoContext
}

// enabled — пишет ли логгер строки запроса на уровне options.level. Логгер
// без Enabled (моки, сторонние реализации) считается пишущим — поведение
// прежнее; *slog.Logger его умеет.
func (o *options) enabled(ctx context.Context, logger logging.Logger) bool {
	checker, ok := logger.(interface {
		Enabled(ctx context.Context, level slog.Level) bool
	})
	if !ok {
		return true
	}

	level := slog.LevelInfo
	if o.level <= logging.Levels.DEBUG {
		level = slog.LevelDebug
	}

	return checker.Enabled(ctx, level)
}
