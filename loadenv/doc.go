// Package loadenv provides easy access to env variables from .env file.
// Also, loadenv provides option to pass default environments to use, if searched variable will not be found.
//
// Lists (GetEnvAsSlice, IsStringIsValidSlice) follow one contract: a single
// element is a valid list, elements are trimmed at both ends, empty elements are
// dropped, and the default value is returned only when nothing is left — the
// variable is unset, empty, or built of separators alone. An empty separator
// also gives the default value: splitting by it would yield a slice of
// characters. The price of trimming is that a value whose edge whitespace is
// significant will not reach the caller intact; write such whitespace as an
// escape (`\s`, `[ ]`) rather than a literal.
package loadenv
