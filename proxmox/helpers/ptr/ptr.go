package ptr

// Ptr creates a ptr from a value to use it inline.
func Ptr[T any](val T) *T {
	return &val
}

// Or will dereference a pointer and return the given value if it's nil.
func Or[T any](p *T, or T) T {
	if p != nil {
		return *p
	}

	return or
}

// func must[T any](val T, err error) T {
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	return val
// }
