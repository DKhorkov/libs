package graphql

import (
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
)

// RequestBody represents information about Request Body for later parsing.
type RequestBody struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// QueryInfo represents information about GraphQL query, it's params and variables.
type QueryInfo struct {
	Type       string            `json:"type"`       // Request type (query, mutation, subscription)
	Name       string            `json:"name"`       // Request name
	Parameters map[string]string `json:"parameters"` // Request params (type of requested variables)
	Variables  map[string]any    `json:"variables"`  // Request variables for query
	Fields     []FieldInfo       `json:"fields"`     // List of requested fields, including input variables
}

// FieldInfo represents information about GraphQL Field and it's arguments.
type FieldInfo struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ParseQuery parses GraphQL-request and returns info about it.
func ParseQuery(query string) (*QueryInfo, error) {
	// Parsing request to AST tree:
	doc, err := parser.Parse(
		parser.ParseParams{
			Source: query,
		},
	)
	if err != nil {
		return nil, &ParseError{BaseErr: err}
	}

	info := &QueryInfo{
		Parameters: make(map[string]string),
		Variables:  make(map[string]any),
		Fields:     []FieldInfo{},
	}

	// Проходим по определениям в документе
	for _, def := range doc.Definitions {
		// Check, if an expression is an operation (query, mutation, subscription):
		if op, ok := def.(*ast.OperationDefinition); ok {
			info.Type = op.Operation
			if op.Name != nil {
				info.Name = op.Name.Value
			}

			// Parsing variables (parameters):
			for _, variable := range op.VariableDefinitions {
				if variable.Variable != nil && variable.Type != nil {
					varName := variable.Variable.Name.Value
					varType := variable.Type.String()
					info.Parameters[varName] = varType
				}
			}

			// Parsing fields, including embedded:
			if op.SelectionSet != nil {
				info.Fields = extractFields(op.SelectionSet)
			}
		}
	}

	// Real name of operation in GraphQL Schema. Last element should be removed from Fields.
	// For example
	//
	// "query { users { user(id: "123") { id name } } }"
	//
	// name of operation is user and users if custom name, which is useless.
	if info.Name == "" {
		info.Name = info.Fields[len(info.Fields)-2].Name
		info.Fields = info.Fields[:len(info.Fields)-1]
	} else {
		info.Name = info.Fields[len(info.Fields)-1].Name
	}

	return info, nil
}

// extractFields recursively extracts all fields and their arguments.
func extractFields(selectionSet *ast.SelectionSet) []FieldInfo {
	var fields []FieldInfo
	if selectionSet == nil {
		return fields
	}

	for _, selection := range selectionSet.Selections {
		if field, ok := selection.(*ast.Field); ok {
			fieldInfo := FieldInfo{
				Name:      field.Name.Value,
				Arguments: extractArguments(field.Arguments),
			}

			// Recursively extracts embedded fields:
			if field.SelectionSet != nil {
				nestedFields := extractFields(field.SelectionSet)
				fields = append(fields, nestedFields...)
			}

			fields = append(fields, fieldInfo)
		}
	}

	return fields
}

// extractArguments extracts field arguments.
func extractArguments(args []*ast.Argument) map[string]any {
	arguments := make(map[string]any)

	for _, arg := range args {
		if arg.Value != nil {
			arguments[arg.Name.Value] = extractValue(arg.Value)
		}
	}

	return arguments
}

// extractValue extracts value AST-node.
func extractValue(value ast.Value) any {
	switch v := value.(type) {
	case *ast.IntValue:
		return v.Value
	case *ast.FloatValue:
		return v.Value
	case *ast.StringValue:
		return v.Value
	case *ast.BooleanValue:
		return v.Value
	case *ast.ObjectValue:
		obj := make(map[string]any)
		for _, field := range v.Fields {
			obj[field.Name.Value] = extractValue(field.Value)
		}

		return obj
	case *ast.ListValue:
		list := make([]any, len(v.Values))
		for i, val := range v.Values {
			list[i] = extractValue(val)
		}

		return list
	default:
		return nil
	}
}
