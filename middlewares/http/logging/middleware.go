package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/middlewares/http/intercepting_response_writer"
)

// truncatedPayload — что пишется в лог вместо тела ответа, не уместившегося в
// перехват.
const truncatedPayload = "<body truncated>"

// redactedValue — что пишется в лог вместо значения учётных данных.
const redactedValue = "<redacted>"

// nonJSONBodyMessage — тело запроса не JSON (загрузка файла): тело в лог не
// пишется. Пишется на уровне строк запроса, а не ошибкой — сбоя нет.
const nonJSONBodyMessage = "Request body is not JSON, payload is not logged"

// serverErrorMessage — строка ERROR на ответ 5xx: пишется на любом уровне
// строк запроса, чтобы упавший запрос был виден в журнале, даже когда строки
// запросов ушли на DEBUG.
const serverErrorMessage = "Request failed"

// errServerErrorResponse — ошибка строки serverErrorMessage: подробности сбоя
// знает хендлер (и трасса), middleware видит только статус.
var errServerErrorResponse = errors.New("server error response")

// setCookieHeader — заголовок ответа, выдающий куку: имя куки в логе
// остаётся, значение и атрибуты — нет.
const setCookieHeader = "Set-Cookie"

// refererHeader — адрес страницы, с которой пришёл запрос: в логе от него
// остаётся только сайт (схема и хост).
const refererHeader = "Referer"

// credentialHeaders — заголовки с учётными данными, канонические имена.
// Значения в лог не пишутся ни из запроса, ни из ответа: строка лога с кукой
// сессии — это вход чужой сессией для любого, у кого есть доступ к журналу.
// X-Csrf-Token — по той же причине: при double-submit его значение равно куке
// csrfToken, и заголовок, записанный как есть, вернул бы спрятанную куку.
var credentialHeaders = map[string]bool{
	"Authorization":       true,
	"Proxy-Authorization": true,
	"Cookie":              true,
	setCookieHeader:       true,
	"X-Csrf-Token":        true,
	refererHeader:         true,
}

// redactHeaders — копия заголовков для лога, в которой значения учётных
// данных заменены на redactedValue. Оригинал не меняется: по нему работают
// хендлер и клиент.
//
// Ключи сравниваются в каноническом виде: хендлер может записать заголовок
// ответа прямо в map (`w.Header()["set-cookie"]`), и поиск по каноническому
// имени такой ключ пропустил бы вместе с токеном.
func redactHeaders(headers http.Header) http.Header {
	redacted := headers.Clone()

	for key, values := range redacted {
		name := http.CanonicalHeaderKey(key)
		if !credentialHeaders[name] {
			continue
		}

		for i, value := range values {
			values[i] = redactedHeaderValue(name, value)
		}
	}

	return redacted
}

// redactedHeaderValue — значение заголовка с учётными данными для лога. У
// Set-Cookie остаётся имя куки (`accessToken=<redacted>`): видно, что ответ
// выдал, но не что. У Referer остаётся сайт: браузер без Referrer-Policy шлёт
// своему же сайту полный адрес страницы, и код из пути страницы
// подтверждения почты или почта из её query ехали в лог с каждым запросом
// за статикой.
func redactedHeaderValue(name, value string) string {
	switch name {
	case setCookieHeader:
		if cookieName, _, found := strings.Cut(value, "="); found {
			return cookieName + "=" + redactedValue
		}
	case refererHeader:
		if referer, err := url.Parse(
			value,
		); err == nil && referer.Scheme != "" &&
			referer.Host != "" {
			return referer.Scheme + "://" + referer.Host
		}
	}

	return redactedValue
}

// cookieNames — имена кук запроса без значений: видно, с чем пришёл запрос,
// а сессию из лога не унести.
func cookieNames(cookies []*http.Cookie) []string {
	names := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		names = append(names, cookie.Name)
	}

	return names
}

// Middleware — New с одними полями тела, которые не пишутся в лог. Прежняя
// сигнатура для существующих вызовов; адрес запроса пишется целиком.
func Middleware(
	logger logging.Logger,
	sensitiveFields ...string,
) func(next http.Handler) http.Handler {
	return New(logger, WithSensitiveFields(sensitiveFields...))
}

// New пишет в лог каждый запрос и ответ: метод, адрес, заголовки, имена кук,
// тело запроса без полей WithSensitiveFields, статус ответа и тело ответа —
// только у ошибки (responsePayload).
//
// Значения учётных данных в лог не попадают: Authorization,
// Proxy-Authorization, Cookie и X-CSRF-Token заменяются на "<redacted>", у
// Set-Cookie остаётся только имя куки, а куки запроса пишутся одними именами.
// Хендлер и клиент видят настоящие значения — редактируется копия заголовков.
// Секреты в адресе запроса закрывает WithRequestURL.
func New(logger logging.Logger, opts ...Option) func(next http.Handler) http.Handler {
	options := newOptions(opts...)
	logLine := options.logLine()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Уровень строк выключен — ни чтения тела, ни копии ответа: загрузка
			// до 20 МиБ читалась в память и тело каждого ответа копировалось ради
			// строки, которую никто не напишет. Статус перехватывается: 5xx
			// пишется на любом уровне.
			if !options.enabled(ctx, logger) {
				rw := intercepting_response_writer.NewStatusOnly(w)
				next.ServeHTTP(rw, r)
				logServerError(ctx, logger, r, rw.StatusCode, options)

				return
			}

			// Reading request body:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logging.LogErrorContext(
					ctx,
					logger,
					"Failed to log request due to reading request body failure",
					err,
				)
			}

			// Restoring request body for later usage due to the fact that io.Reader can be read only once:
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			var payloadInput map[string]any

			if len(body) > 0 {
				if err = json.Unmarshal(body, &payloadInput); err != nil {
					logLine(ctx, logger, nonJSONBodyMessage, "Error", err.Error())
				}

				for _, field := range options.sensitiveFields {
					delete(payloadInput, field)
				}
			}

			// Адрес считается до хендлера: маршрут уже сопоставлен, и обе
			// строки лога пишут один и тот же адрес.
			loggedURL := options.loggedURL(r)

			// Logging request info:
			logLine(
				ctx,
				logger,
				"Received new request",
				options.requestFields(r, loggedURL, payloadInput)...,
			)

			// Create new New for response intercepting purpose:
			rw := intercepting_response_writer.New(w)
			next.ServeHTTP(rw, r)

			// Тело — только у ошибки: успешный ответ несёт данные игрока, а фильтр
			// полей есть только у тела запроса.
			payloadOutput := responsePayload(rw)

			// Logging response info:
			logLine(
				ctx,
				logger,
				"Received response",
				[]any{
					"For", r.Host,
					"Method", r.Method,
					"URL", loggedURL,
					"StatusCode", rw.StatusCode,
					"Headers", redactHeaders(rw.Header()),
					"Payload", payloadOutput,
				}...,
			)

			logServerError(ctx, logger, r, rw.StatusCode, options)
		})
	}
}

// loggedURL — адрес запроса для лога: из WithRequestURL или r.URL целиком.
func (o *options) loggedURL(r *http.Request) any {
	if o.requestURL != nil {
		return o.requestURL(r)
	}

	return r.URL
}

// requestFields — поля строки лога запроса. Без WithRequestURL query пишется
// отдельным полем, как раньше; с ней — только то, что вернула функция
// приложения.
func (o *options) requestFields(r *http.Request, loggedURL any, payload map[string]any) []any {
	fields := []any{
		"From", r.Host,
		"Method", r.Method,
		"URL", loggedURL,
		"Headers", redactHeaders(r.Header),
	}

	if o.requestURL == nil {
		fields = append(fields, "Query", r.URL.Query())
	}

	return append(
		fields,
		"Cookies", cookieNames(r.Cookies()),
		"Form", r.PostForm,
		"Payload", payload,
	)
}

// responsePayload — тело ответа для лога: у ошибки (статус 400 и выше) — её
// текст, у успешного ответа — ничего.
//
// Успешный ответ несёт данные игрока: почту в профиле, код приватного лобби.
// Фильтр полей работает только для тела запроса, и строка ответа выдавала их
// целиком; тело ошибки — её текст, его в лог и нужно.
func responsePayload(rw *intercepting_response_writer.InterceptingResponseWriter) any {
	if rw.StatusCode < http.StatusBadRequest || len(rw.Body) == 0 {
		return nil
	}

	if rw.BodyTruncated {
		return truncatedPayload
	}

	return map[string]string{"error": string(rw.Body)}
}

// logServerError — строка ERROR на ответ 5xx: метод, адрес (WithRequestURL —
// без секретов), статус. Подробности сбоя — в трассе по Request ID.
func logServerError(
	ctx context.Context,
	logger logging.Logger,
	r *http.Request,
	status int,
	options *options,
) {
	if status < http.StatusInternalServerError {
		return
	}

	logging.LogErrorContext(
		ctx,
		logger,
		serverErrorMessage,
		errServerErrorResponse,
		"Method", r.Method,
		"URL", options.loggedURL(r),
		"StatusCode", status,
	)
}
