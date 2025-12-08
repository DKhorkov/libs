package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	graphqlparser "github.com/DKhorkov/libs/graphql"
	"github.com/DKhorkov/libs/logging"
)

const (
	graphqlURLPath = "/query"

	inputFieldName    = "input"
	passwordFieldName = "password"
)

// GraphQLLoggingMiddleware logs GraphQL request info, such as query type, name, fields, variables and return fields.
func GraphQLLoggingMiddleware(next http.Handler, logger logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != graphqlURLPath {
			next.ServeHTTP(w, r)

			return
		}

		ctx := r.Context()

		// Reading request body:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logging.LogErrorContext(
				ctx,
				logger,
				"Failed to log request due to reading request body failure",
				err,
			)

			next.ServeHTTP(w, r)

			return
		}

		// Restoring request body for later usage due to the fact that io.Reader can be read only once:
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		// Parsing request body:
		requestBody := &graphqlparser.RequestBody{}
		if err = json.Unmarshal(body, requestBody); err != nil {
			logging.LogErrorContext(
				ctx,
				logger,
				"Failed to log request due to invalid JSON",
				err,
			)

			next.ServeHTTP(w, r)

			return
		}

		// Retrieving request info:
		info, err := graphqlparser.ParseQuery(requestBody.Query)
		if err != nil {
			logging.LogErrorContext(
				ctx,
				logger,
				"Failed to log request due to GraphQL query parse failure",
				err,
			)

			next.ServeHTTP(w, r)

			return
		}

		// Adding variable to info about request:
		info.Variables = requestBody.Variables
		hideSensitiveInfo(info)

		// Logging request info:
		logging.LogInfoContext(
			ctx,
			logger,
			fmt.Sprintf(
				"Received new request: Type=%s, Name=%s, Parameters=%+v, Variables=%+v, Fields=%+v\n",
				info.Type,
				info.Name,
				info.Parameters,
				info.Variables,
				info.Fields,
			),
		)

		next.ServeHTTP(w, r)
	})
}

// hideSensitiveInfo hides sensitive fields from being logged.
func hideSensitiveInfo(info *graphqlparser.QueryInfo) {
	for index := range info.Fields {
		if _, ok := info.Fields[index].Arguments[inputFieldName]; ok {
			if _, ok = info.Fields[index].Arguments[inputFieldName].(map[string]any); ok {
				if _, ok = info.Fields[index].Arguments[inputFieldName].(map[string]any)[passwordFieldName]; ok {
					info.Fields[index].Arguments[inputFieldName].(map[string]any)[passwordFieldName] = ""
				}
			}
		}
	}

	if _, ok := info.Variables[inputFieldName]; ok {
		if _, ok = info.Variables[inputFieldName].(map[string]any); ok {
			if _, ok = info.Variables[inputFieldName].(map[string]any)[passwordFieldName]; ok {
				info.Variables[inputFieldName].(map[string]any)[passwordFieldName] = ""
			}
		}
	}
}
