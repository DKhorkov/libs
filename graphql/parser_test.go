package graphql_test

import (
	"testing"

	graphqlparser "github.com/DKhorkov/libs/graphql"
	"github.com/stretchr/testify/require"
)

func TestParseQuery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		query         string
		expected      *graphqlparser.QueryInfo
		errorExpected bool
		err           error
	}{
		{
			name:  "mutation success",
			query: "mutation CreateUser { createUser(input: { name: \"John\", age: 30 }) { id name } }",
			expected: &graphqlparser.QueryInfo{
				Name:       "createUser",
				Type:       "mutation",
				Parameters: make(map[string]string),
				Variables:  make(map[string]any),
				Fields: []graphqlparser.FieldInfo{
					{
						Name:      "id",
						Arguments: make(map[string]any),
					},
					{
						Name:      "name",
						Arguments: make(map[string]any),
					},
					{
						Name: "createUser",
						Arguments: map[string]any{
							"input": map[string]any{"name": "John", "age": "30"},
						},
					},
				},
			},
		},
		{
			name:  "query success",
			query: "query GetUser { user(id: \"123\") { id name } }",
			expected: &graphqlparser.QueryInfo{
				Name:       "user",
				Type:       "query",
				Parameters: make(map[string]string),
				Variables:  make(map[string]any),
				Fields: []graphqlparser.FieldInfo{
					{
						Name:      "id",
						Arguments: make(map[string]any),
					},
					{
						Name:      "name",
						Arguments: make(map[string]any),
					},
					{
						Name:      "user",
						Arguments: map[string]any{"id": "123"},
					},
				},
			},
		},

		// Кривой кейс для парсинга. Сложно спарсить название нормально с сильно вложенной структурой.
		{
			name:  "query user",
			query: "query { GetUser { user(id: \"123\") { id name } } }",
			expected: &graphqlparser.QueryInfo{
				Name:       "GetUser",
				Type:       "query",
				Parameters: make(map[string]string),
				Variables:  make(map[string]any),
				Fields: []graphqlparser.FieldInfo{
					{
						Name:      "id",
						Arguments: make(map[string]any),
					},
					{
						Name:      "name",
						Arguments: make(map[string]any),
					},
					{
						Name:      "user",
						Arguments: map[string]any{"id": "123"},
					},
					{
						Name:      "GetUser",
						Arguments: make(map[string]any),
					},
				},
			},
		},
		{
			name:          "invalid request expression",
			query:         "invalid",
			errorExpected: true,
			err:           &graphqlparser.ParseError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := graphqlparser.ParseQuery(tc.query)
			if tc.errorExpected {
				require.Error(t, err)
				require.IsType(t, tc.err, err)
			}

			require.Equal(t, tc.expected, actual)
		})
	}
}
