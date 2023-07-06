package ptr

// Ptr creates a ptr from a value to use it inline.
func Ptr[T any](val T) *T {
	return &val
}
