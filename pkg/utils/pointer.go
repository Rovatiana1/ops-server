package utils

// Ptr returns a pointer to the given value. Useful for optional fields.
func Ptr[T any](v T) *T {
	return &v
}

// Val dereferences a pointer, returning the zero value if nil.
func Val[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
