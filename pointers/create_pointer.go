package pointers

// New create new pointer to provided value.
func New[T any](value T) *T {
	return &value
}
